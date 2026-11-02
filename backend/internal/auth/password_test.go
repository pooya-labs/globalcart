package auth

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

func TestHashPasswordAndVerify(t *testing.T) {
	password := "correct horse battery staple — نمونه"
	encoded, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Fatalf("unexpected PHC parameters: %q", encoded)
	}
	if strings.Contains(encoded, password) {
		t.Fatal("PHC output contains the plaintext password")
	}

	ok, err := VerifyPassword(encoded, password)
	if err != nil || !ok {
		t.Fatalf("VerifyPassword(correct) = (%t, %v), want (true, nil)", ok, err)
	}
	ok, err = VerifyPassword(encoded, "incorrect password")
	if err != nil || ok {
		t.Fatalf("VerifyPassword(wrong) = (%t, %v), want (false, nil)", ok, err)
	}
	needsRehash, err := PasswordHashNeedsRehash(encoded)
	if err != nil || needsRehash {
		t.Fatalf("PasswordHashNeedsRehash(current) = (%t, %v), want (false, nil)", needsRehash, err)
	}
}

func TestHashPasswordRejectsInvalidInput(t *testing.T) {
	for _, password := range []string{"", strings.Repeat("x", maxPasswordBytes+1), string([]byte{0xff, 0xfe})} {
		if _, err := HashPassword(password); !errors.Is(err, ErrInvalidPassword) {
			t.Errorf("HashPassword(%q) error = %v, want ErrInvalidPassword", password, err)
		}
	}
}

func TestVerifyPasswordRejectsMalformedOrOverBudgetPHC(t *testing.T) {
	salt := base64.RawStdEncoding.EncodeToString(make([]byte, saltLength))
	key := base64.RawStdEncoding.EncodeToString(make([]byte, keyLength))
	cases := []string{
		"",
		"$argon2i$v=19$m=19456,t=2,p=1$" + salt + "$" + key,
		"$argon2id$v=16$m=19456,t=2,p=1$" + salt + "$" + key,
		"$argon2id$v=19$m=1,t=2,p=1$" + salt + "$" + key,
		"$argon2id$v=19$m=999999,t=2,p=1$" + salt + "$" + key,
		"$argon2id$v=19$m=19456,t=99,p=1$" + salt + "$" + key,
		"$argon2id$v=19$m=19456,t=2,p=99$" + salt + "$" + key,
		"$argon2id$v=19$m=19456,t=2,t=3,p=1$" + salt + "$" + key,
		"$argon2id$v=19$m=19456,t=2,p=1$" + salt + "$",
		"$argon2id$v=19$m=19456,t=2,p=1$" + salt + "$" + key + "$extra",
		"$argon2id$v=19$m=19456,t=2,p=1$not-base64$" + key,
		"$argon2id$v=19$m=19456,t=2,p=1$" + salt + "$" + base64.RawStdEncoding.EncodeToString(make([]byte, 8)),
	}
	for _, encoded := range cases {
		if ok, err := VerifyPassword(encoded, "password"); ok || !errors.Is(err, ErrInvalidHash) {
			t.Errorf("VerifyPassword(%q) = (%t, %v), want (false, ErrInvalidHash)", encoded, ok, err)
		}
	}
}

func TestPasswordHashNeedsRehashRejectsWeakHash(t *testing.T) {
	salt := base64.RawStdEncoding.EncodeToString(make([]byte, saltLength))
	key := base64.RawStdEncoding.EncodeToString(make([]byte, keyLength))
	weak := "$argon2id$v=19$m=12288,t=1,p=1$" + salt + "$" + key
	if needs, err := PasswordHashNeedsRehash(weak); needs || !errors.Is(err, ErrInvalidHash) {
		t.Fatalf("PasswordHashNeedsRehash(weak) = (%t, %v), want (false, ErrInvalidHash)", needs, err)
	}
}

func TestVerifyPasswordRejectsInvalidPasswordInput(t *testing.T) {
	if ok, err := VerifyPassword("not-a-hash", ""); ok || !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("VerifyPassword(empty) = (%t, %v), want ErrInvalidPassword", ok, err)
	}
}
