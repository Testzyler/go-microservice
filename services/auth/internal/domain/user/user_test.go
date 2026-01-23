package user

import "testing"

func TestNormalizeEmail(t *testing.T) {
	got, err := NormalizeEmail(" User@Example.COM ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != "user@example.com" {
		t.Fatalf("expected normalized email, got %s", got)
	}
}

func TestNormalizeEmailInvalid(t *testing.T) {
	if _, err := NormalizeEmail("invalid"); err == nil {
		t.Fatalf("expected error for invalid email")
	}
}

func TestValidatePassword(t *testing.T) {
	if err := ValidatePassword("short"); err == nil {
		t.Fatalf("expected error for short password")
	}
	if err := ValidatePassword("longenoughpassword"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
