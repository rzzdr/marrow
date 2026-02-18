package format

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

func MarshalYAMLString(v any) (string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}
