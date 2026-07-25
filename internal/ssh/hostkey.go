package sshclient

import (
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
)

func hostKeyFingerprint(key ssh.PublicKey) string {
	return ssh.FingerprintSHA256(key)
}

func verifyHostKey(expected string, pin func(fingerprint string) error) ssh.HostKeyCallback {
	return func(_ string, _ net.Addr, key ssh.PublicKey) error {
		fingerprint := hostKeyFingerprint(key)
		if expected == "" {
			if pin == nil {
				return fmt.Errorf("ssh host key not pinned and no pin callback configured")
			}
			if err := pin(fingerprint); err != nil {
				return fmt.Errorf("pin host key: %w", err)
			}
			return nil
		}
		if fingerprint != expected {
			return fmt.Errorf("ssh host key mismatch: got %s, expected %s", fingerprint, expected)
		}
		return nil
	}
}
