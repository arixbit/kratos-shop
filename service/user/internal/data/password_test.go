package data

import (
	"strings"
	"testing"
)

func TestEncryptPassword(t *testing.T) {
	hash := EncryptPassword("admin123")
	parts := strings.Split(hash, "$")
	if len(parts) != 4 {
		t.Fatalf("unexpected hash format: %q", hash)
	}
	if parts[1] != "pbkdf2-sha512" {
		t.Fatalf("unexpected algorithm: %q", parts[1])
	}
	if parts[2] == "" || parts[3] == "" {
		t.Fatalf("salt or encoded password is empty: %q", hash)
	}
	if hash == EncryptPassword("admin123") {
		t.Fatal("same password should produce different salt each time")
	}
}

func TestVerifyPassword(t *testing.T) {
	hash := EncryptPassword("admin123")

	ok, err := VerifyPassword("admin123", hash)
	if err != nil {
		t.Fatalf("verify correct password: %v", err)
	}
	if !ok {
		t.Fatal("correct password should verify")
	}

	ok, err = VerifyPassword("wrong", hash)
	if err != nil {
		t.Fatalf("verify wrong password: %v", err)
	}
	if ok {
		t.Fatal("wrong password should not verify")
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	cases := []string{
		"",
		"plaintext",
		"$pbkdf2-sha512$salt",
		"$unknown$salt$encoded",
	}
	for _, c := range cases {
		if _, err := VerifyPassword("admin123", c); err == nil {
			t.Fatalf("expected error for malformed hash %q", c)
		}
	}
}
