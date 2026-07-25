package crypto

import "testing"

func TestEncryptDecrypt_roundtrip(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	ct, err := Encrypt([]byte("secret"), key)
	if err != nil {
		t.Fatal(err)
	}
	pt, err := Decrypt(ct, key)
	if err != nil {
		t.Fatal(err)
	}
	if string(pt) != "secret" {
		t.Fatalf("got %q", pt)
	}
}
