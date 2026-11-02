package database

import (
	"crypto/sha256"
	"encoding/hex"
)

func migrationChecksum(contents []byte) string {
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}
