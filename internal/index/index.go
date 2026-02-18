package index

import (
	"github.com/rzzdr/marrow/internal/model"
	"github.com/rzzdr/marrow/internal/store"
)

func Rebuild(s *store.Store) (model.Index, error) {
	idx, _ := s.ReadIndex()

	proj, err := s.ReadProject()
	if err != nil {
		return idx, err
	}

	exps, err := s.ListExperiments()
	if err != nil {
		return idx, err
	}

	learnings, err := s.ReadLearnings()
	if err != nil {
		return idx, err
	}

	graveyard, err := s.ReadGraveyard()
	if err != nil {
		return idx, err
	}

	idx.Computed = Compute(exps, learnings, graveyard, proj.Metric)

	if err := s.WriteIndex(idx); err != nil {
		return idx, err
	}

	return idx, nil
}

func UpdateIncremental(s *store.Store, newExp model.Experiment) (model.Index, error) {
	idx, err := s.ReadIndex()
	if err != nil {
		return Rebuild(s)
	}

	proj, err := s.ReadProject()
	if err != nil {
		return idx, err
	}

	c := &idx.Computed
	if c.StatusCounts == nil {
		c.StatusCounts = make(map[string]int)
	}
	c.TotalExperiments++
	c.StatusCounts[newExp.Status]++

	tagSet := make(map[string]bool)
	for _, t := range c.AllTags {
		tagSet[t] = true
	}
	for _, t := range newExp.Tags {
		if !tagSet[t] {
			c.AllTags = append(c.AllTags, t)
			tagSet[t] = true
		}
	}

	isBetter := false
	if c.BestMetric == nil {
		isBetter = true
	} else {
		higher := proj.Metric.Direction == "higher_is_better"
		if higher {
			isBetter = newExp.Metric.Value > c.BestMetric.Value
		} else {
			isBetter = newExp.Metric.Value < c.BestMetric.Value
		}
	}

	if isBetter {
		c.BestExperiment = newExp.ID
		c.BestMetric = &newExp.Metric

		exps, err := s.ListExperiments()
		if err == nil {
			learnings, _ := s.ReadLearnings()
			graveyard, _ := s.ReadGraveyard()
			full := Compute(exps, learnings, graveyard, proj.Metric)
			c.ExperimentChain = full.ExperimentChain
		}
	}

	if err := s.WriteIndex(idx); err != nil {
		return idx, err
	}
	return idx, nil
}
