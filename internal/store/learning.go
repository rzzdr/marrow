package store

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rzzdr/marrow/internal/format"
	"github.com/rzzdr/marrow/internal/model"
)

func (s *Store) ReadLearnings() (model.LearningsFile, error) {
	var lf model.LearningsFile
	err := format.ReadYAML(s.learningsPath(), &lf)
	return lf, err
}

func (s *Store) WriteLearnings(lf model.LearningsFile) error {
	return format.WriteYAML(s.learningsPath(), lf)
}

func (s *Store) AddLearning(l model.Learning) (string, error) {
	lf, err := s.ReadLearnings()
	if err != nil {
		return "", fmt.Errorf("reading learnings: %w", err)
	}

	maxNum := 0
	for _, existing := range lf.Proven {
		if n := parseLearningNum(existing.ID); n > maxNum {
			maxNum = n
		}
	}
	for _, existing := range lf.Assumptions {
		if n := parseLearningNum(existing.ID); n > maxNum {
			maxNum = n
		}
	}
	l.ID = fmt.Sprintf("learn_%0*d", idWidth(maxNum+1), maxNum+1)
	if l.Timestamp.IsZero() {
		l.Timestamp = time.Now().UTC()
	}

	switch l.Type {
	case model.LearningProven:
		lf.Proven = append(lf.Proven, l)
	case model.LearningAssumption:
		lf.Assumptions = append(lf.Assumptions, l)
	default:
		lf.Assumptions = append(lf.Assumptions, l)
	}

	if err := s.WriteLearnings(lf); err != nil {
		return "", err
	}
	return l.ID, nil
}

func (s *Store) ReadGraveyard() (model.GraveyardFile, error) {
	var gf model.GraveyardFile
	err := format.ReadYAML(s.graveyardPath(), &gf)
	return gf, err
}

func (s *Store) WriteGraveyard(gf model.GraveyardFile) error {
	return format.WriteYAML(s.graveyardPath(), gf)
}

func (s *Store) AddGraveyardEntry(g model.GraveyardEntry) (string, error) {
	gf, err := s.ReadGraveyard()
	if err != nil {
		return "", fmt.Errorf("reading graveyard: %w", err)
	}

	maxNum := 0
	for _, existing := range gf.Entries {
		if n := parseGraveyardNum(existing.ID); n > maxNum {
			maxNum = n
		}
	}
	g.ID = fmt.Sprintf("grave_%0*d", idWidth(maxNum+1), maxNum+1)
	if g.Timestamp.IsZero() {
		g.Timestamp = time.Now().UTC()
	}

	gf.Entries = append(gf.Entries, g)

	if err := s.WriteGraveyard(gf); err != nil {
		return "", err
	}
	return g.ID, nil
}

func (s *Store) DeleteLearning(id string) error {
	lf, err := s.ReadLearnings()
	if err != nil {
		return err
	}

	found := false
	var newProven []model.Learning
	for _, l := range lf.Proven {
		if l.ID == id {
			found = true
			continue
		}
		newProven = append(newProven, l)
	}

	var newAssumptions []model.Learning
	for _, l := range lf.Assumptions {
		if l.ID == id {
			found = true
			continue
		}
		newAssumptions = append(newAssumptions, l)
	}

	if !found {
		return fmt.Errorf("learning %s not found", id)
	}

	lf.Proven = newProven
	lf.Assumptions = newAssumptions
	return s.WriteLearnings(lf)
}

func (s *Store) DeleteGraveyardEntry(id string) error {
	gf, err := s.ReadGraveyard()
	if err != nil {
		return err
	}

	found := false
	var remaining []model.GraveyardEntry
	for _, g := range gf.Entries {
		if g.ID == id {
			found = true
			continue
		}
		remaining = append(remaining, g)
	}

	if !found {
		return fmt.Errorf("graveyard entry %s not found", id)
	}

	gf.Entries = remaining
	return s.WriteGraveyard(gf)
}

func parseLearningNum(id string) int {
	numStr := strings.TrimPrefix(id, "learn_")
	n, _ := strconv.Atoi(numStr)
	return n
}

func parseGraveyardNum(id string) int {
	numStr := strings.TrimPrefix(id, "grave_")
	n, _ := strconv.Atoi(numStr)
	return n
}

func idWidth(n int) int {
	if n < 1000 {
		return 3
	}
	return len(strconv.Itoa(n))
}
