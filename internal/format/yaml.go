package format

import (
	"bytes"
	"os"

	"gopkg.in/yaml.v3"
)

func ReadYAML(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, target)
}

func WriteYAML(path string, source any) error {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(source); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
