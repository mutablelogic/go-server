package types

import (
	"crypto/sha256"
	"encoding/hex"
)

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func Hash(data []byte) string {
	// Return the SHA256 hash of the data
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
