package htpasswd

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	// Modules
	"github.com/GehirnInc/crypt/apr1_crypt"
	"golang.org/x/crypto/bcrypt"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Htpasswd struct {
	passwords map[string]string
}

type HashAlgorithm int

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	passwordSeparator = ":"
	prefixSHA         = "{SHA}"
	prefixBCrypt      = "$apr1$"
	prefixMD5a        = "$2a$"
	prefixMD5y        = "$2y$"
)

const (
	BCrypt HashAlgorithm = iota
	MD5
	SHA
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func New() *Htpasswd {
	this := new(Htpasswd)
	this.passwords = make(map[string]string)
	return this
}

func Read(r io.Reader) (*Htpasswd, error) {
	this := New()

	// Scan the file
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line == "" {
			continue
		} else if strings.HasPrefix(line, "#") {
			continue
		} else if passwd := strings.SplitN(line, passwordSeparator, 2); len(passwd) != 2 {
			continue
		} else {
			this.passwords[passwd[0]] = passwd[1]
		}
	}

	// Return any errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Return success
	return this, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *Htpasswd) String() string {
	str := "<htpasswd"
	if users := this.Users(); len(users) > 0 {
		str += fmt.Sprintf(" users=%q", users)
	}
	return str + ">"
}

func (a HashAlgorithm) String() string {
	switch a {
	case BCrypt:
		return "BCrypt"
	case MD5:
		return "MD5"
	case SHA:
		return "SHA"
	default:
		return "[?? Invalid HashAlgorithm value]"
	}
}

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

// Return with usernames
func (this *Htpasswd) Users() []string {
	var users []string
	for user := range this.passwords {
		users = append(users, user)
	}
	return users
}

// Set a password for a user with a named hashing algorithm
func (this *Htpasswd) Set(name, passwd string, hash HashAlgorithm) error {
	if name == "" {
		return ErrBadParameter.With("name")
	}
	if passwd == "" {
		return ErrBadParameter.With("passwd")
	}
	switch hash {
	case BCrypt:
		if passwd, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost); err != nil {
			return err
		} else {
			this.passwords[name] = string(passwd)
		}
	case MD5:
		if passwd, err := apr1_crypt.New().Generate([]byte(passwd), nil); err != nil {
			return err
		} else {
			this.passwords[name] = passwd
		}
	case SHA:
		this.passwords[name] = "{SHA}" + base64.StdEncoding.EncodeToString([]byte(passwd))
	default:
		return ErrBadParameter.With(hash)
	}

	// Return success
	return nil
}

// Delete a password by user
func (this *Htpasswd) Delete(name string) {
	delete(this.passwords, name)
}

// Validate a password for a user
func (this *Htpasswd) Verify(name, passwd string) bool {
	hash, exists := this.passwords[name]
	if !exists {
		return false
	}
	switch {
	case strings.HasPrefix(hash, prefixSHA):
		digest := strings.TrimPrefix(hash, prefixSHA)
		if digest, err := base64.StdEncoding.DecodeString(digest); err != nil {
			return false
		} else {
			return string(digest) == passwd
		}
	case strings.HasPrefix(hash, prefixMD5a):
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passwd)); err == nil {
			return true
		}
	case strings.HasPrefix(hash, prefixMD5y):
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passwd)); err == nil {
			return true
		}
	case strings.HasPrefix(hash, prefixBCrypt):
		if err := apr1_crypt.New().Verify(hash, []byte(passwd)); err == nil {
			return true
		}
	}
	return false
}

// Write out passwords
func (this *Htpasswd) Write(w io.Writer) error {
	for user, password := range this.passwords {
		if _, err := fmt.Fprintln(w, user+passwordSeparator+password); err != nil {
			return err
		}
	}
	return nil
}
