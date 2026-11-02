package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
)

const opaqueTokenBytes = 32

var ErrInvalidOpaqueToken = errors.New("invalid opaque token")

// NewOpaqueToken returns 256 bits of cryptographically secure randomness in
// canonical unpadded base64url form. Store only DigestOpaqueToken(token).
func NewOpaqueToken() (string, error) {
	raw := make([]byte, opaqueTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	defer clear(raw)
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// DigestOpaqueToken validates a canonical 256-bit base64url token and returns
// its SHA-256 digest for persistence. The raw token must never be stored or logged.
func DigestOpaqueToken(token string) ([32]byte, error) {
	var zero [32]byte
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != opaqueTokenBytes ||
		base64.RawURLEncoding.EncodeToString(raw) != token {
		clear(raw)
		return zero, ErrInvalidOpaqueToken
	}
	defer clear(raw)
	return sha256.Sum256(raw), nil
}

// EqualTokenDigests compares stored token digests without data-dependent timing.
func EqualTokenDigests(left, right [32]byte) bool {
	return subtle.ConstantTimeCompare(left[:], right[:]) == 1
}
