package format

import (
	"bytes"
	"os"
	"path/filepath"

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

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".marrow-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	mode := os.FileMode(0644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode()
	}

	if _, err := tmp.Write(buf.Bytes()); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}
