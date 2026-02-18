package model

import "time"

type ChangelogEntry struct {
	Timestamp time.Time `yaml:"ts"`
	Action    string    `yaml:"action"`            // exp_logged | learning_added | graveyard_added | index_rebuilt | pinned_updated | snapshot_created | context_updated
	ID        string    `yaml:"id,omitempty"`      // relevant entity ID
	Type      string    `yaml:"type,omitempty"`    // sub-type (e.g. proven, assumption)
	Summary   string    `yaml:"summary,omitempty"` // human-readable one-liner
}

type ChangelogFile struct {
	Entries []ChangelogEntry `yaml:"entries"`
}
