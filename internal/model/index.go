package model

import "time"

type Index struct {
	Computed ComputedIndex `yaml:"computed"`
	Pinned   PinnedIndex   `yaml:"pinned"`
}

type ComputedIndex struct {
	LastUpdated      time.Time      `yaml:"last_updated"`
	TotalExperiments int            `yaml:"total_experiments"`
	BestExperiment   string         `yaml:"best_experiment,omitempty"`
	BestMetric       *MetricResult  `yaml:"best_metric,omitempty"`
	ExperimentChain  []string       `yaml:"experiment_chain,omitempty"` // best path through the DAG
	AllTags          []string       `yaml:"all_tags,omitempty"`
	StatusCounts     map[string]int `yaml:"status_counts,omitempty"`
	ProvenCount      int            `yaml:"proven_count"`
	AssumptionCount  int            `yaml:"assumption_count"`
	GraveyardCount   int            `yaml:"graveyard_count"`
}

type PinnedIndex struct {
	DoNotTry         []string `yaml:"do_not_try,omitempty"`
	Deferred         []string `yaml:"deferred,omitempty"`
	DataWarnings     []string `yaml:"data_warnings,omitempty"`
	CriticalFeatures []string `yaml:"critical_features,omitempty"`
	Notes            string   `yaml:"notes,omitempty"`
}
