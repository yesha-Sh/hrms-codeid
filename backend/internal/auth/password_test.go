package auth

import "testing"

func TestPasswordHasherHashAndVerify(t *testing.T) {
	hasher := NewPasswordHasher()

	hash, err := hasher.Hash("ChangeMe123!")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	ok, err := hasher.Verify("ChangeMe123!", hash)
	if err != nil {
		t.Fatalf("verify valid password: %v", err)
	}
	if !ok {
		t.Fatal("expected valid password verification to succeed")
	}

	ok, err = hasher.Verify("WrongPassword!", hash)
	if err != nil {
		t.Fatalf("verify invalid password: %v", err)
	}
	if ok {
		t.Fatal("expected invalid password verification to fail")
	}
}
