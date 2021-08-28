package main

import (
	"math/rand"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// CONSTANTS

const (
	SECRET_KEY_LENGTH = 20
)

const (
	secretKeyString = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func NewSecret() []byte {
	return randStringBytes(SECRET_KEY_LENGTH)
}

func isValidSecret(value []byte) bool {
	if len(value) != SECRET_KEY_LENGTH {
		return false
	}
	// TODO: Actually compare with secretKeyString
	return true
}

// randStringBytes returns a new secret key with n characters
func randStringBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = secretKeyString[rand.Intn(len(secretKeyString))]
	}
	return b
}
