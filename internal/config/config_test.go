package config

import "testing"

func TestLoad_requiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL missing")
	}
}

func TestLoad_ok(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/otelforge")
	t.Setenv("ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.WorkerConcurrency < 1 {
		t.Fatal("expected positive worker concurrency")
	}
}
