package format

import (
	"github.com/rzzdr/marrow/internal/model"
)

func FilterExperiment(e model.Experiment, depth model.Depth) model.Experiment {
	switch depth {
	case model.DepthSummary:
		return model.Experiment{
			ID:     e.ID,
			Status: e.Status,
			Metric: e.Metric,
			Tags:   e.Tags,
		}
	case model.DepthStandard:
		return model.Experiment{
			ID:          e.ID,
			Timestamp:   e.Timestamp,
			BaseModel:   e.BaseModel,
			Parents:     e.Parents,
			ChangesFrom: e.ChangesFrom,
			Metric:      e.Metric,
			Status:      e.Status,
			LocalCV:     e.LocalCV,
			PublicLB:    e.PublicLB,
			DataVersion: e.DataVersion,
			Tags:        e.Tags,
			Notes:       e.Notes,
		}
	default:
		return e
	}
}

func FilterLearning(l model.Learning, depth model.Depth) model.Learning {
	switch depth {
	case model.DepthSummary:
		return model.Learning{
			ID:   l.ID,
			Type: l.Type,
			Text: l.Text,
		}
	case model.DepthStandard:
		return model.Learning{
			ID:        l.ID,
			Timestamp: l.Timestamp,
			Type:      l.Type,
			Text:      l.Text,
			Tags:      l.Tags,
		}
	default:
		return l
	}
}

func EstimateTokens(s string) int {
	return (len(s) + 3) / 4
}
