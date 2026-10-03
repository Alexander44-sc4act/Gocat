package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKeyRejectsEmptySalt(t *testing.T) {
	if _, err := DeriveKey("password", nil, 32); err == nil {
		t.Fatal("DeriveKey(nil salt) = nil error, want rejection")
	}
	if _, err := DeriveKey("password", []byte{}, 32); err == nil {
		t.Fatal("DeriveKey(empty salt) = nil error, want rejection")
	}
}

func TestDeriveKeyRandomSaltDiffers(t *testing.T) {
	s1, err := GenerateKey(32)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	s2, err := GenerateKey(32)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	k1, err := DeriveKey("password", s1, 32)
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}
	k2, err := DeriveKey("password", s2, 32)
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}
	if len(k1) != 32 || len(k2) != 32 {
		t.Fatalf("key lengths = %d, %d, want 32, 32", len(k1), len(k2))
	}
	if bytes.Equal(k1, k2) {
		t.Fatal("same password with different salts derived identical keys")
	}
}
