package store

import (
	"fmt"
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

	total := len(lf.Proven) + len(lf.Assumptions)
	l.ID = fmt.Sprintf("learn_%03d", total+1)
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

	g.ID = fmt.Sprintf("grave_%03d", len(gf.Entries)+1)
	if g.Timestamp.IsZero() {
		g.Timestamp = time.Now().UTC()
	}

	gf.Entries = append(gf.Entries, g)

	if err := s.WriteGraveyard(gf); err != nil {
		return "", err
	}
	return g.ID, nil
}
