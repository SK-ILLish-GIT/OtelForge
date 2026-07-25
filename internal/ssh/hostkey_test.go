package sshclient

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestVerifyHostKeyPinsOnFirstConnect(t *testing.T) {
	key := testPublicKey(t)
	var pinned string
	cb := verifyHostKey("", func(fp string) error {
		pinned = fp
		return nil
	})
	if err := cb("host", nil, key); err != nil {
		t.Fatalf("first connect: %v", err)
	}
	if pinned == "" {
		t.Fatal("expected fingerprint to be pinned")
	}
}

func TestVerifyHostKeyRejectsMismatch(t *testing.T) {
	key := testPublicKey(t)
	cb := verifyHostKey("SHA256:wrong", nil)
	if err := cb("host", nil, key); err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestVerifyHostKeyAcceptsMatch(t *testing.T) {
	key := testPublicKey(t)
	fp := hostKeyFingerprint(key)
	cb := verifyHostKey(fp, nil)
	if err := cb("host", nil, key); err != nil {
		t.Fatalf("matching fingerprint: %v", err)
	}
}

func testPublicKey(t *testing.T) ssh.PublicKey {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return pub
}
