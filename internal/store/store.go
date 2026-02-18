package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rzzdr/marrow/internal/format"
	"github.com/rzzdr/marrow/internal/model"
)

const marrowDir = ".marrow"

type Store struct {
	root string // absolute path to the .marrow/ directory
}

func New(projectDir string) *Store {
	return &Store{root: filepath.Join(projectDir, marrowDir)}
}

func (s *Store) Root() string {
	return s.root
}

func (s *Store) Exists() bool {
	info, err := os.Stat(s.root)
	return err == nil && info.IsDir()
}

func (s *Store) Init(project model.Project) error {
	dirs := []string{
		s.root,
		filepath.Join(s.root, "experiments"),
		filepath.Join(s.root, "learnings"),
		filepath.Join(s.root, "context"),
		filepath.Join(s.root, "snapshots"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	if err := format.WriteYAML(s.projectPath(), project); err != nil {
		return fmt.Errorf("writing project config: %w", err)
	}

	idx := model.Index{}
	if err := format.WriteYAML(s.indexPath(), idx); err != nil {
		return fmt.Errorf("writing index: %w", err)
	}

	cl := model.ChangelogFile{}
	if err := format.WriteYAML(s.changelogPath(), cl); err != nil {
		return fmt.Errorf("writing changelog: %w", err)
	}

	if err := format.WriteYAML(s.learningsPath(), model.LearningsFile{}); err != nil {
		return fmt.Errorf("writing learnings: %w", err)
	}
	if err := format.WriteYAML(s.graveyardPath(), model.GraveyardFile{}); err != nil {
		return fmt.Errorf("writing graveyard: %w", err)
	}

	return nil
}

func (s *Store) projectPath() string {
	return filepath.Join(s.root, "marrow.yaml")
}

func (s *Store) indexPath() string {
	return filepath.Join(s.root, "index.yaml")
}

func (s *Store) changelogPath() string {
	return filepath.Join(s.root, "changelog.yaml")
}

func (s *Store) experimentsDir() string {
	return filepath.Join(s.root, "experiments")
}

func (s *Store) experimentPath(id string) string {
	return filepath.Join(s.root, "experiments", id+".yaml")
}

func (s *Store) learningsPath() string {
	return filepath.Join(s.root, "learnings", "learnings.yaml")
}

func (s *Store) graveyardPath() string {
	return filepath.Join(s.root, "learnings", "graveyard.yaml")
}

func (s *Store) contextDir() string {
	return filepath.Join(s.root, "context")
}

func (s *Store) contextPath(name string) string {
	return filepath.Join(s.root, "context", name+".yaml")
}

func (s *Store) snapshotsDir() string {
	return filepath.Join(s.root, "snapshots")
}
