package model

import "fmt"

type Project struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Template    string            `yaml:"template,omitempty"`
	TaskType    string            `yaml:"task_type,omitempty"`
	Metric      MetricDef         `yaml:"metric"`
	DataVersion int               `yaml:"data_version,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Extra       map[string]string `yaml:"extra,omitempty"`
}

type MetricDef struct {
	Name      string  `yaml:"name"`
	Direction string  `yaml:"direction"`
	Baseline  float64 `yaml:"baseline,omitempty"`
}

func (m MetricDef) Validate() error {
	switch m.Direction {
	case "higher_is_better", "lower_is_better":
		return nil
	default:
		return fmt.Errorf("invalid metric direction %q: must be higher_is_better or lower_is_better", m.Direction)
	}
}
