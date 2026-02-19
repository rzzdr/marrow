package store

import (
	"fmt"
	"os"
	"time"

	"github.com/rzzdr/marrow/internal/format"
	"github.com/rzzdr/marrow/internal/model"
)

func (s *Store) ReadChangelog() (model.ChangelogFile, error) {
	var cf model.ChangelogFile
	err := format.ReadYAML(s.changelogPath(), &cf)
	return cf, err
}

const maxChangelogEntries = 1000

func (s *Store) AppendChangelog(entry model.ChangelogEntry) error {
	cf, err := s.ReadChangelog()
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("reading changelog: %w", err)
		}
		cf = model.ChangelogFile{}
	}

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	cf.Entries = append(cf.Entries, entry)

	// Rotate: keep only the most recent entries
	if len(cf.Entries) > maxChangelogEntries {
		cf.Entries = cf.Entries[len(cf.Entries)-maxChangelogEntries:]
	}

	return format.WriteYAML(s.changelogPath(), cf)
}

func (s *Store) ReadChangelogSince(since time.Time) ([]model.ChangelogEntry, error) {
	cf, err := s.ReadChangelog()
	if err != nil {
		return nil, err
	}

	var filtered []model.ChangelogEntry
	for _, e := range cf.Entries {
		if !e.Timestamp.Before(since) {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil
}
