package store

import (
	"github.com/rzzdr/marrow/internal/format"
	"github.com/rzzdr/marrow/internal/model"
)

func (s *Store) ReadProject() (model.Project, error) {
	var p model.Project
	err := format.ReadYAML(s.projectPath(), &p)
	return p, err
}

func (s *Store) WriteProject(p model.Project) error {
	return format.WriteYAML(s.projectPath(), p)
}

func (s *Store) ReadIndex() (model.Index, error) {
	var idx model.Index
	err := format.ReadYAML(s.indexPath(), &idx)
	return idx, err
}

func (s *Store) WriteIndex(idx model.Index) error {
	return format.WriteYAML(s.indexPath(), idx)
}
