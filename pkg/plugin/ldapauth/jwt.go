package main

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	// Modules
	"github.com/dgrijalva/jwt-go"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type JWT struct {
	sync.Mutex

	secret []byte
	valid  time.Time
	fn     SecretFunc
	expiry time.Duration
}

type Claims struct {
	Values url.Values
	jwt.StandardClaims
}

type SecretFunc func(context.Context) ([]byte, error)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	// DefaultExpiry is how long to keep the existing token for before it
	// neeeds re-validated, not the expiry time for a user session
	DefaultExpiry = time.Minute
)

var (
	SigningMethod = jwt.SigningMethodHS256
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func NewJWT(fn SecretFunc) *JWT {
	return NewJWTWithExpiry(fn, DefaultExpiry)
}

func NewJWTWithExpiry(fn SecretFunc, expiry time.Duration) *JWT {
	this := new(JWT)
	this.fn = fn
	this.expiry = expiry
	return this
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func (this *JWT) Token(ctx context.Context, values url.Values) (string, time.Time, error) {
	now := time.Now()
	expiry := now.Add(this.expiry)
	token := jwt.NewWithClaims(SigningMethod, &Claims{
		Values: values,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			ExpiresAt: expiry.Unix(),
		},
	})
	if secret, err := this.getSecret(ctx); err != nil {
		return "", expiry, err
	} else if value, err := token.SignedString(secret); err != nil {
		return "", expiry, err
	} else {
		return value, expiry, nil
	}
}

func (this *JWT) Parse(ctx context.Context, token string) (url.Values, time.Time, error) {
	claims := &Claims{}
	if token, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return this.getSecret(ctx)
	}); isErrorInvalidToken(err) {
		return nil, time.Time{}, ErrInvalidCredentials
	} else if isErrorExpiredToken(err) {
		// Return the expired token so that it might be renewed
		return claims.Values, time.Unix(claims.ExpiresAt, 0), ErrExpiredCredentials
	} else if err != nil {
		return nil, time.Time{}, err
	} else if token.Valid == false {
		return nil, time.Time{}, ErrInvalidCredentials
	}

	// Success
	return claims.Values, time.Unix(claims.ExpiresAt, 0), nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *JWT) String() string {
	str := "<ldapauth.jwt"
	str += fmt.Sprintf(" expiry=%v", this.expiry)
	return str + ">"
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func (this *JWT) getSecret(ctx context.Context) ([]byte, error) {
	this.Mutex.Lock()
	defer this.Mutex.Unlock()

	// Obtain the current secret from provider
	var now = time.Now()
	if this.valid.IsZero() || this.secret == nil || now.After(this.valid) {
		if secret, err := this.fn(ctx); err != nil {
			return nil, err
		} else {
			this.secret = secret
			this.valid = now.Add(this.expiry)
		}
	}

	// Return the secret
	return this.secret, nil
}

func isErrorExpiredToken(err error) bool {
	if err == nil {
		return false
	}
	if err, ok := err.(*jwt.ValidationError); ok {
		if err.Errors&jwt.ValidationErrorExpired != 0 {
			return true
		}
	}
	return false
}

func isErrorInvalidToken(err error) bool {
	if err == nil {
		return false
	}
	if err, ok := err.(*jwt.ValidationError); ok {
		if err.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
			return true
		}
	}
	return false
}
