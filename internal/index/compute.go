package index

import (
	"sort"
	"strings"
	"time"

	"github.com/rzzdr/marrow/internal/model"
)

func Compute(
	exps []model.Experiment,
	learnings model.LearningsFile,
	graveyard model.GraveyardFile,
	metric model.MetricDef,
) model.ComputedIndex {
	ci := model.ComputedIndex{
		LastUpdated:      time.Now().UTC(),
		TotalExperiments: len(exps),
		ProvenCount:      len(learnings.Proven),
		AssumptionCount:  len(learnings.Assumptions),
		GraveyardCount:   len(graveyard.Entries),
		StatusCounts:     make(map[string]int),
	}

	if len(exps) == 0 {
		return ci
	}

	tagSet := make(map[string]bool)
	for _, e := range exps {
		ci.StatusCounts[e.Status]++
		for _, t := range e.Tags {
			tagSet[t] = true
		}
	}
	for t := range tagSet {
		ci.AllTags = append(ci.AllTags, t)
	}
	sort.Strings(ci.AllTags)

	best := findBest(exps, metric)
	if best != nil {
		ci.BestExperiment = best.ID
		ci.BestMetric = &best.Metric
	}

	if best != nil {
		ci.ExperimentChain = computeChain(exps, *best, metric)
	}

	return ci
}

func findBest(exps []model.Experiment, metric model.MetricDef) *model.Experiment {
	if len(exps) == 0 {
		return nil
	}

	higher := strings.EqualFold(metric.Direction, "higher_is_better")
	var best *model.Experiment
	for i := range exps {
		if exps[i].Status == "failed" {
			continue
		}
		if best == nil {
			best = &exps[i]
			continue
		}
		if higher {
			if exps[i].Metric.Value > best.Metric.Value {
				best = &exps[i]
			}
		} else {
			if exps[i].Metric.Value < best.Metric.Value {
				best = &exps[i]
			}
		}
	}
	return best
}

func computeChain(exps []model.Experiment, best model.Experiment, metric model.MetricDef) []string {
	expMap := make(map[string]model.Experiment, len(exps))
	for _, e := range exps {
		expMap[e.ID] = e
	}

	higher := strings.EqualFold(metric.Direction, "higher_is_better")

	var chain []string
	current := best
	visited := make(map[string]bool)

	for {
		chain = append(chain, current.ID)
		visited[current.ID] = true

		if len(current.Parents) == 0 {
			break
		}

		var bestParent *model.Experiment
		for _, pid := range current.Parents {
			if visited[pid] {
				continue
			}
			if p, ok := expMap[pid]; ok {
				if bestParent == nil {
					bestParent = &p
				} else if higher && p.Metric.Value > bestParent.Metric.Value {
					bestParent = &p
				} else if !higher && p.Metric.Value < bestParent.Metric.Value {
					bestParent = &p
				}
			}
		}

		if bestParent == nil {
			break
		}
		current = *bestParent
	}

	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}
