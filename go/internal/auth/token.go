package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

const tokenBytes = 32

var (
	ErrMalformedToken = errors.New("malformed session token")
	ErrBadSignature   = errors.New("session token signature mismatch")
)

var enc = base64.RawURLEncoding

func NewToken() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}

	return enc.EncodeToString(buf), nil
}

func SignToken(token, secret string) string {
	return token + "." + enc.EncodeToString(sign(token, secret))
}

func UnsignToken(signed, secret string) (string, error) {
	token, sig, ok := strings.Cut(signed, ".")
	if !ok || token == "" || sig == "" {
		return "", ErrMalformedToken
	}

	got, err := enc.DecodeString(sig)
	if err != nil {
		return "", ErrMalformedToken
	}

	if !hmac.Equal(got, sign(token, secret)) {
		return "", ErrBadSignature
	}

	return token, nil
}

func sign(token, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(token))
	return mac.Sum(nil)
}
