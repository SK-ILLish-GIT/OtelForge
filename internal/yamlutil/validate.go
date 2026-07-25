package yamlutil

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func ValidateYAML(content []byte) error {
	var doc any
	if err := yaml.Unmarshal(content, &doc); err != nil {
		return fmt.Errorf("invalid yaml: %w", err)
	}
	return nil
}
