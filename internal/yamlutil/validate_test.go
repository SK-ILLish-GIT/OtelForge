package yamlutil

import "testing"

func TestValidateYAML_rejectsInvalid(t *testing.T) {
	err := ValidateYAML([]byte(":\n  bad"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateYAML_acceptsValid(t *testing.T) {
	err := ValidateYAML([]byte("receivers:\n  otlp:\n    protocols:\n      grpc:\n"))
	if err != nil {
		t.Fatal(err)
	}
}
