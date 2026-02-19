package util

import (
	"fmt"
	"path/filepath"
	"strings"
)

func SafeName(name string) error {
	cleaned := filepath.Clean(name)
	if cleaned != name || strings.ContainsAny(name, `/\`) || name == ".." || name == "." || name == "" {
		return fmt.Errorf("invalid name %q: must be a plain filename without path separators", name)
	}
	return nil
}