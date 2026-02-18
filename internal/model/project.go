package model

type Project struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Template    string            `yaml:"template,omitempty"`
	TaskType    string            `yaml:"task_type,omitempty"` // classification, regression, generation, etc.
	Metric      MetricDef         `yaml:"metric"`              // primary evaluation metric
	DataVersion int               `yaml:"data_version,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Extra       map[string]string `yaml:"extra,omitempty"` // arbitrary key-value pairs
}

type MetricDef struct {
	Name      string  `yaml:"name"`               // e.g. AUC-ROC, RMSE, BLEU
	Direction string  `yaml:"direction"`          // higher_is_better | lower_is_better
	Baseline  float64 `yaml:"baseline,omitempty"` // baseline value for comparison
}
