package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

func parseKey(key string) ([]byte, error) {
	if len(key) == 32 {
		return []byte(key), nil
	}
	if len(key) == 64 {
		b, err := hex.DecodeString(key)
		if err != nil {
			return nil, fmt.Errorf("invalid hex ENCRYPTION_KEY: %w", err)
		}
		if len(b) != 32 {
			return nil, fmt.Errorf("ENCRYPTION_KEY must decode to 32 bytes")
		}
		return b, nil
	}
	return nil, fmt.Errorf("ENCRYPTION_KEY must be 32 bytes raw or 64 hex chars")
}

func Encrypt(plaintext []byte, keyStr string) ([]byte, error) {
	key, err := parseKey(keyStr)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(ciphertext []byte, keyStr string) ([]byte, error) {
	key, err := parseKey(keyStr)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ct := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ct, nil)
}
