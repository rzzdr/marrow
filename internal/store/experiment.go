package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/rzzdr/marrow/internal/format"
	"github.com/rzzdr/marrow/internal/model"
)

func (s *Store) NextExperimentID() (string, error) {
	entries, err := os.ReadDir(s.experimentsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return "exp_001", nil
		}
		return "", err
	}

	maxNum := 0
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".yaml")
		if strings.HasPrefix(name, "exp_") {
			numStr := strings.TrimPrefix(name, "exp_")
			if n, err := strconv.Atoi(numStr); err == nil && n > maxNum {
				maxNum = n
			}
		}
	}
	width := 3
	if maxNum >= 999 {
		width = len(strconv.Itoa(maxNum + 1))
	}
	return fmt.Sprintf("exp_%0*d", width, maxNum+1), nil
}

func (s *Store) WriteExperiment(exp model.Experiment) error {
	return format.WriteYAML(s.experimentPath(exp.ID), exp)
}

func (s *Store) ReadExperiment(id string) (model.Experiment, error) {
	var exp model.Experiment
	err := format.ReadYAML(s.experimentPath(id), &exp)
	return exp, err
}

func (s *Store) ListExperiments() ([]model.Experiment, error) {
	entries, err := os.ReadDir(s.experimentsDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var exps []model.Experiment
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		var exp model.Experiment
		if err := format.ReadYAML(filepath.Join(s.experimentsDir(), e.Name()), &exp); err != nil {
			return nil, fmt.Errorf("reading %s: %w", e.Name(), err)
		}
		exps = append(exps, exp)
	}

	sort.Slice(exps, func(i, j int) bool {
		return exps[i].ID < exps[j].ID
	})
	return exps, nil
}

func (s *Store) ListExperimentsByTag(tags []string) ([]model.Experiment, error) {
	all, err := s.ListExperiments()
	if err != nil {
		return nil, err
	}

	tagSet := make(map[string]bool, len(tags))
	for _, t := range tags {
		tagSet[t] = true
	}

	var filtered []model.Experiment
	for _, exp := range all {
		for _, t := range exp.Tags {
			if tagSet[t] {
				filtered = append(filtered, exp)
				break
			}
		}
	}
	return filtered, nil
}

func (s *Store) DeleteExperiment(id string) error {
	path := s.experimentPath(id)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("experiment %s not found", id)
	}
	return os.Remove(path)
}

func (s *Store) FindParentRefs(id string) ([]string, error) {
	exps, err := s.ListExperiments()
	if err != nil {
		return nil, err
	}
	var refs []string
	for _, e := range exps {
		for _, pid := range e.Parents {
			if pid == id {
				refs = append(refs, e.ID)
				break
			}
		}
	}
	return refs, nil
}
