package auth

import "testing"

const secret = "test-secret"

func TestSignUnsignRoundTrip(t *testing.T) {
	token, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}

	got, err := UnsignToken(SignToken(token, secret), secret)
	if err != nil {
		t.Fatalf("UnsignToken: %v", err)
	}
	if got != token {
		t.Fatalf("round trip changed the token: %q != %q", got, token)
	}
}

func TestNewTokenIsUnique(t *testing.T) {
	seen := make(map[string]bool, 128)
	for range 128 {
		token, err := NewToken()
		if err != nil {
			t.Fatalf("NewToken: %v", err)
		}
		if seen[token] {
			t.Fatal("NewToken returned a duplicate")
		}
		seen[token] = true
	}
}

func TestUnsignTokenRejectsTamperedValues(t *testing.T) {
	token, err := NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	signed := SignToken(token, secret)

	tests := map[string]struct {
		value   string
		secret  string
		wantErr error
	}{
		"empty":            {"", secret, ErrMalformedToken},
		"no signature":     {token, secret, ErrMalformedToken},
		"empty signature":  {token + ".", secret, ErrMalformedToken},
		"not base64":       {token + ".!!!!", secret, ErrMalformedToken},
		"swapped payload":  {"forged." + signed[len(token)+1:], secret, ErrBadSignature},
		"different secret": {signed, "other-secret", ErrBadSignature},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := UnsignToken(tc.value, tc.secret); err != tc.wantErr {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}
