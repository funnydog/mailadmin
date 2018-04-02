package sha512crypt

import (
	"bytes"
	//	"fmt"
	"testing"
)

func TestSha512cryptingIsEasy(t *testing.T) {
	pass := []byte("mypassword")
	hp, err := GenerateFromPassword(pass)
	if err != nil {
		t.Fatalf("GenerateFromPassword error: %s", err)
	}

	if CompareHashAndPassword(hp, pass) != nil {
		t.Errorf("%v should hash %s correctly", hp, pass)
	}

	notPass := "notthepass"
	err = CompareHashAndPassword(hp, []byte(notPass))
	if err != ErrMismatchedHashAndPassword {
		t.Errorf("%v and %s should be mismatched", hp, notPass)
	}
}

func TestSha512cryptingIsCorrect(t *testing.T) {
	pass := []byte("allmine")
	salt := []byte("bpKfXE60l8fkqZtO")
	expectedHash := []byte("$6$bpKfXE60l8fkqZtO$Um5EJkQJMczwkJU5Flonw7n244dJIhTbYOgu507juLnN3H/i433cd0i/uP25Bfw9m0Ce7PqgmY93JxYhC1Lp1.")

	hash, err := sha512crypt(pass, salt, defaultRounds)
	if err != nil {
		t.Fatalf("sha512crypt blew up: %v", err)
	}

	if !bytes.HasSuffix(expectedHash, hash) {
		t.Errorf("%v should be the suffix of %v", string(hash), string(expectedHash))
	}

	h, err := newFromHash(expectedHash)
	if err != nil {
		t.Errorf("Unable to parse %s: %v", string(expectedHash), err)
	}

	if err == nil && !bytes.Equal(expectedHash, h.Hash()) {
		t.Errorf("Parsed hash %v should equal %v", h.Hash(), expectedHash)
	}
}

func TestVeryShortPasswords(t *testing.T) {
	key := []byte("k")
	salt := []byte("bpKfXE60l8fkqZtO")
	_, err := sha512crypt(key, salt, defaultRounds)
	if err != nil {
		t.Errorf("One byte key resulted in error: %s", err)
	}
}
