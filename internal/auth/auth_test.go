package auth

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHashRoundTrip(t *testing.T) {
	password := "test-password-123"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		t.Fatalf("hashing password: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		t.Errorf("password should match hash: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte("wrong-password")); err == nil {
		t.Error("wrong password should not match hash")
	}
}

func TestGenerateTokenUniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generateToken()
		if err != nil {
			t.Fatalf("generating token: %v", err)
		}
		if len(token) != 64 {
			t.Errorf("token length = %d, want 64", len(token))
		}
		if tokens[token] {
			t.Errorf("duplicate token generated: %s", token)
		}
		tokens[token] = true
	}
}
