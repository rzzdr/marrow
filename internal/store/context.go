package store

import (
	"fmt"
	"os"
	"strings"

	"github.com/rzzdr/marrow/internal/format"
	"github.com/rzzdr/marrow/internal/util"
)

func (s *Store) ReadContext(name string) (map[string]any, error) {
	if err := util.SafeName(name); err != nil {
		return nil, err
	}
	var data map[string]any
	err := format.ReadYAML(s.contextPath(name), &data)
	return data, err
}

func (s *Store) WriteContext(name string, data map[string]any) error {
	if err := util.SafeName(name); err != nil {
		return err
	}
	return format.WriteYAML(s.contextPath(name), data)
}

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
