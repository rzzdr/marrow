package model

import "time"

type Experiment struct {
	ID        string    `yaml:"id"`
	Timestamp time.Time `yaml:"timestamp"`
	BaseModel string    `yaml:"base_model,omitempty"` // model family: xgboost, resnet, llama, etc.

	Parents     []string            `yaml:"parents,omitempty"`
	ChangesFrom map[string][]Change `yaml:"changes_from,omitempty"` // parent_id → list of changes

	Metric    MetricResult `yaml:"metric"`
	Status    string       `yaml:"status"` // improved | degraded | neutral | failed
	Reasoning Reasoning    `yaml:"reasoning,omitempty"`

	Environment *Environment `yaml:"environment,omitempty"`

	LocalCV  *float64 `yaml:"local_cv,omitempty"`
	PublicLB *float64 `yaml:"public_lb,omitempty"`

	DataVersion int `yaml:"data_version,omitempty"`

	Tags  []string `yaml:"tags,omitempty"`
	Notes string   `yaml:"notes,omitempty"`
}

type Change struct {
	Type  string `yaml:"type,omitempty"`  // param | added | removed | changed
	Param string `yaml:"param,omitempty"` // parameter name if type=param
	What  string `yaml:"what,omitempty"`  // description if type=added/removed/changed
	From  string `yaml:"from,omitempty"`
	To    string `yaml:"to,omitempty"`
}

type MetricResult struct {
	Name     string  `yaml:"name"`
	Value    float64 `yaml:"value"`
	Baseline float64 `yaml:"baseline,omitempty"`
	Delta    float64 `yaml:"delta,omitempty"`
}

type Reasoning struct {
	Type     string            `yaml:"type"` // proven | assumption | unknown
	Text     string            `yaml:"text"`
	Evidence map[string]string `yaml:"evidence,omitempty"` // exp_id → observation
}

type Environment struct {
	Python            string            `yaml:"python,omitempty"`
	GPU               string            `yaml:"gpu,omitempty"`
	KeyPackages       map[string]string `yaml:"key_packages,omitempty"`
	DataHash          string            `yaml:"data_hash,omitempty"`
	SplitSeed         *int              `yaml:"split_seed,omitempty"`
	PreprocessingHash string            `yaml:"preprocessing_hash,omitempty"`
}
