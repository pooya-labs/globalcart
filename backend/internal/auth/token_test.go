package auth

import (
	"encoding/base64"
	"testing"
)

func TestNewOpaqueTokenAndDigest(t *testing.T) {
	token, err := NewOpaqueToken()
	if err != nil {
		t.Fatalf("NewOpaqueToken: %v", err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("decode token: %v", err)
	}
	if len(raw) != opaqueTokenBytes {
		t.Fatalf("decoded token length = %d, want %d", len(raw), opaqueTokenBytes)
	}
	digest, err := DigestOpaqueToken(token)
	if err != nil {
		t.Fatalf("DigestOpaqueToken: %v", err)
	}
	if EqualTokenDigests(digest, [32]byte{}) {
		t.Fatal("token digest is zero")
	}
	again, err := DigestOpaqueToken(token)
	if err != nil || !EqualTokenDigests(digest, again) {
		t.Fatalf("same token produced inconsistent digest: equal=%t err=%v", EqualTokenDigests(digest, again), err)
	}
	otherToken, err := NewOpaqueToken()
	if err != nil {
		t.Fatalf("NewOpaqueToken(second): %v", err)
	}
	otherDigest, err := DigestOpaqueToken(otherToken)
	if err != nil {
		t.Fatalf("DigestOpaqueToken(second): %v", err)
	}
	if EqualTokenDigests(digest, otherDigest) {
		t.Fatal("independent opaque tokens produced the same digest")
	}
}

func TestDigestOpaqueTokenRejectsInvalidTokens(t *testing.T) {
	tooShort := base64.RawURLEncoding.EncodeToString(make([]byte, opaqueTokenBytes-1))
	tooLong := base64.RawURLEncoding.EncodeToString(make([]byte, opaqueTokenBytes+1))
	for _, token := range []string{"", "not base64!", tooShort, tooLong, " " + base64.RawURLEncoding.EncodeToString(make([]byte, opaqueTokenBytes))} {
		if digest, err := DigestOpaqueToken(token); err == nil || digest != ([32]byte{}) {
			t.Errorf("DigestOpaqueToken(%q) = (%x, %v), want zero digest and error", token, digest, err)
		}
	}
}
