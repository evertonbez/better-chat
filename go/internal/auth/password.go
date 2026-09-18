package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32
	saltLen             = 16

	MinPasswordLength = 8

	MaxPasswordLength = 128
)

var (
	ErrPasswordTooShort = fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	ErrPasswordTooLong  = fmt.Errorf("password must be at most %d characters", MaxPasswordLength)
	ErrInvalidHash      = errors.New("stored password hash is malformed")
)

const (
	phcFormat = "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
	phcScan   = "$argon2id$v=%d$m=%d,t=%d,p=%d$%s"
)

var b64 = base64.RawStdEncoding

func ValidatePassword(password string) error {
	switch {
	case len(password) < MinPasswordLength:
		return ErrPasswordTooShort
	case len(password) > MaxPasswordLength:
		return ErrPasswordTooLong
	default:
		return nil
	}
}

func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	return fmt.Sprintf(phcFormat,
		argon2.Version, argonMemory, argonTime, argonThreads,
		b64.EncodeToString(salt), b64.EncodeToString(key),
	), nil
}

func VerifyPassword(encoded, password string) (bool, error) {
	if len(password) > MaxPasswordLength {
		return false, ErrPasswordTooLong
	}

	var (
		version            int
		memory, iterations uint32
		threads            uint8
		saltAndKey         string
	)

	if n, err := fmt.Sscanf(encoded, phcScan, &version, &memory, &iterations, &threads, &saltAndKey); n != 5 || err != nil {
		return false, ErrInvalidHash
	}

	if version != argon2.Version {
		return false, ErrInvalidHash
	}

	saltB64, keyB64, ok := strings.Cut(saltAndKey, "$")
	if !ok {
		return false, ErrInvalidHash
	}

	salt, err := b64.DecodeString(saltB64)
	if err != nil {
		return false, ErrInvalidHash
	}

	want, err := b64.DecodeString(keyB64)
	if err != nil || len(want) == 0 {
		return false, ErrInvalidHash
	}

	got := argon2.IDKey([]byte(password), salt, iterations, memory, threads, uint32(len(want)))

	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
