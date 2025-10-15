package configuration

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads the YAML config from a file if present.
// If the file path is empty, it just returns a zero Config (so env-only mode works).
func LoadConfig(path string, out *Config) error {
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true) // catch unknown keys early
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("unmarshalling YAML: %w", err)
	}

	return nil
}
