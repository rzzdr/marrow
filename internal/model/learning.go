package model

import "time"

type LearningType string

const (
	LearningProven     LearningType = "proven"
	LearningAssumption LearningType = "assumption"
	LearningUnknown    LearningType = "unknown"
)

type Learning struct {
	ID        string            `yaml:"id"`
	Timestamp time.Time         `yaml:"timestamp"`
	Type      LearningType      `yaml:"type"`
	Text      string            `yaml:"text"`
	Evidence  map[string]string `yaml:"evidence,omitempty"` // exp_id → observation
	Tags      []string          `yaml:"tags,omitempty"`
}

type GraveyardEntry struct {
	ID           string    `yaml:"id"`
	Timestamp    time.Time `yaml:"timestamp"`
	Approach     string    `yaml:"approach"`
	Reason       string    `yaml:"reason"`
	ExperimentID string    `yaml:"experiment_id,omitempty"` // which experiment proved it failed
	Tags         []string  `yaml:"tags,omitempty"`
}

type LearningsFile struct {
	Proven      []Learning `yaml:"proven,omitempty"`
	Assumptions []Learning `yaml:"assumptions,omitempty"`
}

type GraveyardFile struct {
	Entries []GraveyardEntry `yaml:"entries"`
}
