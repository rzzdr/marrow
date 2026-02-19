package store

import (
	"fmt"
	"os"
	"strings"

	"github.com/rzzdr/marrow/internal/util"
)

func (s *Store) ListContextFiles() ([]string, error) {
	entries, err := os.ReadDir(s.contextDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".yaml") {
			names = append(names, strings.TrimSuffix(e.Name(), ".yaml"))
		}
	}
	return names, nil
}

func (s *Store) ReadContextRaw(name string) (string, error) {
	if err := util.SafeName(name); err != nil {
		return "", err
	}
	data, err := os.ReadFile(s.contextPath(name))
	if err != nil {
		return "", fmt.Errorf("reading context %s: %w", name, err)
	}
	return string(data), nil
}
