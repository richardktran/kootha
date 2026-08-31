package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadYAML decodes a YAML file into dest. Path is relative to the process working directory.
func LoadYAML(path string, dest any) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config %s: %w", path, err)
	}
	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(dest); err != nil {
		return fmt.Errorf("decode config %s: %w", path, err)
	}
	return nil
}
