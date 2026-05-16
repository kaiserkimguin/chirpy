package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if hash == "" {
		t.Fatal("expected hash to not be empty")
	}

	if hash == "password123" {
		t.Fatal("expected hash to be different from password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "password123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !match {
		t.Fatal("expected password to match hash")
	}
}

func TestCheckPasswordHashWrongPassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	match, err := CheckPasswordHash("wrong-password", hash)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if match {
		t.Fatal("expected wrong password not to match hash")
	}
}

func TestGetBearerToken(t *testing.T) {
	req, err := http.NewRequest("GET", "someURL", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	req.Header.Set("Authorization", "Bearer   someRandomString  ")
	token, err := GetBearerToken(req.Header)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token != "someRandomString" {
		t.Fatal("tokens not matching as expected")
	}
}

func TestHappyJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "secret_a"
	expiresIn := time.Hour
	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// reverseCheck
	userIDReverse, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if userID != userIDReverse {
		t.Fatal("expected userID and reverse to match")
	}
	// expired tokeken Test
	expiresInB := time.Second * -1
	tokenStringB, err := MakeJWT(userID, tokenSecret, expiresInB)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_, err = ValidateJWT(tokenStringB, tokenSecret)
	if err == nil {
		t.Fatal("expected validation to fail due to expired token")
	}
	// false secret test
	tokenSecretB := "secret_b"
	_, err = ValidateJWT(tokenString, tokenSecretB)
	if err == nil {
		t.Fatal("expected validation to fail due to wrong secret")
	}
}
