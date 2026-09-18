package auth

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
)

func TestHashPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("super-secret-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	ok, err := VerifyPassword(hash, "super-secret-password")
	if err != nil || !ok {
		t.Fatalf("correct password rejected: ok=%v err=%v", ok, err)
	}

	ok, err = VerifyPassword(hash, "super-secret-passwore")
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if ok {
		t.Fatal("wrong password accepted")
	}
}

func TestHashPasswordEmitsPHCString(t *testing.T) {
	hash, err := HashPassword("super-secret-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	prefix := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$", argon2.Version, argonMemory, argonTime, argonThreads)
	if !strings.HasPrefix(hash, prefix) {
		t.Fatalf("hash %q does not start with %q", hash, prefix)
	}

	if n := strings.Count(hash, "$"); n != 5 {
		t.Fatalf("hash %q has %d fields, want the 5 of a PHC string", hash, n)
	}
}

func TestVerifyPasswordUsesStoredParameters(t *testing.T) {
	const password = "super-secret-password"

	salt := []byte("0123456789abcdef")
	var (
		oldTime    uint32 = 1
		oldMemory  uint32 = 8 * 1024
		oldThreads uint8  = 1
	)
	key := argon2.IDKey([]byte(password), salt, oldTime, oldMemory, oldThreads, 32)
	legacy := fmt.Sprintf(phcFormat, argon2.Version, oldMemory, oldTime, oldThreads,
		b64.EncodeToString(salt), b64.EncodeToString(key))

	ok, err := VerifyPassword(legacy, password)
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if !ok {
		t.Fatal("a hash stored under an older cost was rejected")
	}
}

func TestHashPasswordUsesFreshSalt(t *testing.T) {
	a, err := HashPassword("super-secret-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	b, err := HashPassword("super-secret-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if a == b {
		t.Fatal("the same password hashed twice produced the same digest")
	}
}

func TestValidatePasswordBounds(t *testing.T) {
	if err := ValidatePassword("short"); err == nil {
		t.Fatal("expected a short password to be rejected")
	}
	if err := ValidatePassword(strings.Repeat("a", MaxPasswordLength+1)); err == nil {
		t.Fatal("expected an over-long password to be rejected")
	}
	if err := ValidatePassword(strings.Repeat("a", MinPasswordLength)); err != nil {
		t.Fatalf("expected a password at the minimum length to pass: %v", err)
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	valid, err := HashPassword("super-secret-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	salt, key, _ := strings.Cut(strings.TrimPrefix(valid, "$argon2id$v=19$m=65536,t=3,p=4$"), "$")

	hashes := map[string]string{
		"empty":              "",
		"not a phc string":   "salt:key",
		"scrypt era hash":    "0123456789abcdef:deadbeef",
		"wrong algorithm":    "$argon2i$v=19$m=65536,t=3,p=4$" + salt + "$" + key,
		"unknown version":    "$argon2id$v=18$m=65536,t=3,p=4$" + salt + "$" + key,
		"missing key":        "$argon2id$v=19$m=65536,t=3,p=4$" + salt,
		"salt not base64":    "$argon2id$v=19$m=65536,t=3,p=4$!!!!$" + key,
		"key not base64":     "$argon2id$v=19$m=65536,t=3,p=4$" + salt + "$!!!!",
		"missing parameters": "$argon2id$v=19$" + salt + "$" + key,
	}

	for name, hash := range hashes {
		t.Run(name, func(t *testing.T) {
			ok, err := VerifyPassword(hash, "super-secret-password")
			if ok {
				t.Fatal("a malformed hash was accepted")
			}
			if err != ErrInvalidHash {
				t.Fatalf("expected ErrInvalidHash, got %v", err)
			}
		})
	}
}
