package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

const (
	argonVersion       = argon2.Version
	defaultMemoryKiB   = uint32(19 * 1024)
	defaultIterations  = uint32(2)
	defaultParallelism = uint8(1)
	saltLength         = 16
	keyLength          = 32
	maxPasswordBytes   = 1024

	// Verification accepts stronger hashes up to bounded costs. These limits
	// prevent a corrupted database row from turning login into an unbounded
	// memory/CPU allocation before rate limiting is available.
	maxVerifyMemoryKiB  = uint32(64 * 1024)
	maxVerifyIterations = uint32(5)
	maxVerifyParallel   = uint8(4)
)

var (
	ErrInvalidPassword = errors.New("password must be valid UTF-8 and contain 1 to 1024 bytes")
	ErrInvalidHash     = errors.New("invalid or unsupported Argon2id PHC hash")
)

type phcParameters struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	salt        []byte
	key         []byte
}

// HashPassword returns a PHC-encoded Argon2id hash with a fresh random salt.
// Parameters use OWASP's current baseline; benchmark them on the approved
// deployment runtime before enabling authentication.
func HashPassword(password string) (string, error) {
	if !validPasswordInput(password) {
		return "", ErrInvalidPassword
	}

	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	defer clear(salt)

	key := argon2.IDKey(
		[]byte(password),
		salt,
		defaultIterations,
		defaultMemoryKiB,
		defaultParallelism,
		keyLength,
	)
	defer clear(key)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argonVersion,
		defaultMemoryKiB,
		defaultIterations,
		defaultParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// VerifyPassword validates a PHC hash before deriving a key, enforces strict
// parameter bounds, and compares the derived key in constant time.
func VerifyPassword(encodedHash, password string) (bool, error) {
	if !validPasswordInput(password) {
		return false, ErrInvalidPassword
	}
	params, err := parsePHC(encodedHash)
	if err != nil {
		return false, err
	}
	defer clear(params.salt)
	defer clear(params.key)

	actual := argon2.IDKey(
		[]byte(password),
		params.salt,
		params.iterations,
		params.memory,
		params.parallelism,
		uint32(len(params.key)),
	)
	defer clear(actual)

	return subtle.ConstantTimeCompare(actual, params.key) == 1, nil
}

// PasswordHashNeedsRehash reports whether an otherwise valid hash falls below
// the current local baseline. Stronger hashes are kept rather than downgraded.
func PasswordHashNeedsRehash(encodedHash string) (bool, error) {
	params, err := parsePHC(encodedHash)
	if err != nil {
		return false, err
	}
	defer clear(params.salt)
	defer clear(params.key)

	return params.memory < defaultMemoryKiB ||
		params.iterations < defaultIterations ||
		len(params.salt) < saltLength ||
		len(params.key) < keyLength, nil
}

func validPasswordInput(password string) bool {
	return len(password) > 0 && len(password) <= maxPasswordBytes && utf8.ValidString(password)
}

func parsePHC(encoded string) (phcParameters, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return phcParameters{}, ErrInvalidHash
	}

	version, err := parseUint(strings.TrimPrefix(parts[2], "v="), 32)
	if err != nil || parts[2] != "v="+strconv.FormatUint(version, 10) || version != argonVersion {
		return phcParameters{}, ErrInvalidHash
	}

	values := make(map[string]uint64, 3)
	for _, field := range strings.Split(parts[3], ",") {
		pair := strings.SplitN(field, "=", 2)
		if len(pair) != 2 {
			return phcParameters{}, ErrInvalidHash
		}
		if _, exists := values[pair[0]]; exists {
			return phcParameters{}, ErrInvalidHash
		}
		value, err := parseUint(pair[1], 32)
		if err != nil || strconv.FormatUint(value, 10) != pair[1] {
			return phcParameters{}, ErrInvalidHash
		}
		values[pair[0]] = value
	}
	if len(values) != 3 {
		return phcParameters{}, ErrInvalidHash
	}
	memory, hasMemory := values["m"]
	iterations, hasIterations := values["t"]
	parallelism, hasParallelism := values["p"]
	if !hasMemory || !hasIterations || !hasParallelism ||
		memory < 8*parallelism || memory > uint64(maxVerifyMemoryKiB) ||
		iterations == 0 || iterations > uint64(maxVerifyIterations) ||
		parallelism == 0 || parallelism > uint64(maxVerifyParallel) {
		return phcParameters{}, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 8 || len(salt) > 64 ||
		base64.RawStdEncoding.EncodeToString(salt) != parts[4] {
		clear(salt)
		return phcParameters{}, ErrInvalidHash
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(key) < keyLength/2 || len(key) > 64 ||
		base64.RawStdEncoding.EncodeToString(key) != parts[5] {
		clear(salt)
		clear(key)
		return phcParameters{}, ErrInvalidHash
	}

	return phcParameters{
		memory:      uint32(memory),
		iterations:  uint32(iterations),
		parallelism: uint8(parallelism),
		salt:        salt,
		key:         key,
	}, nil
}

func parseUint(raw string, bitSize int) (uint64, error) {
	if raw == "" {
		return 0, ErrInvalidHash
	}
	value, err := strconv.ParseUint(raw, 10, bitSize)
	if err != nil {
		return 0, ErrInvalidHash
	}
	return value, nil
}
