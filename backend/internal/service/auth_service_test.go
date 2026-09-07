package service

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "SuperSecretPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword() returned an empty hash")
	}

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("expected Argon2id hash, got %q", hash)
	}

	if hash == password {
		t.Fatal("password must not be stored in plain text")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "SuperSecretPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}

	if !valid {
		t.Fatal("expected password to be valid")
	}
}

func TestVerifyPassword_InvalidPassword(t *testing.T) {
	password := "SuperSecretPassword123!"
	wrongPassword := "WrongPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	valid, err := VerifyPassword(wrongPassword, hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}

	if valid {
		t.Fatal("expected password to be invalid")
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "SuperSecretPassword123!"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash1 == hash2 {
		t.Fatal("expected different hashes because salts should be random")
	}
}
