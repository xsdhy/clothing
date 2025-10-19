package auth

import "testing"

func TestPasswordHashingLifecycle(t *testing.T) {
	password := "S3curePass!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}
	if hash == "" {
		t.Fatal("expected hash to be populated")
	}

	if err := VerifyPassword(hash, password); err != nil {
		t.Fatalf("expected password to verify, got error: %v", err)
	}

	if err := VerifyPassword(hash, "wrong"); err == nil {
		t.Fatal("expected verification to fail for wrong password")
	}
}
