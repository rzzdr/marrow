package format

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rzzdr/marrow/internal/model"
)

func ExperimentOneLiner(e model.Experiment) string {
	var parts []string

	parentIDs := make([]string, 0, len(e.ChangesFrom))
	for pid := range e.ChangesFrom {
		parentIDs = append(parentIDs, pid)
	}
	sort.Strings(parentIDs)

	for _, pid := range parentIDs {
		for _, c := range e.ChangesFrom[pid] {
			switch c.Type {
			case "param":
				parts = append(parts, fmt.Sprintf("%s=%s", c.Param, c.To))
			case "added":
				parts = append(parts, "+"+c.What)
			case "removed":
				parts = append(parts, "-"+c.What)
			default:
				if c.What != "" {
					parts = append(parts, c.What)
				}
			}
		}
	}

	changeSummary := strings.Join(parts, ", ")
	if changeSummary == "" && e.BaseModel != "" {
		changeSummary = "baseline " + e.BaseModel
	}

	metricStr := fmt.Sprintf("%s %.4f", e.Metric.Name, e.Metric.Value)
	if e.Metric.Delta != 0 {
		metricStr += fmt.Sprintf(" (%+.4f)", e.Metric.Delta)
	}

	if changeSummary != "" {
		return fmt.Sprintf("%s → %s, %s, %s", e.ID, changeSummary, metricStr, e.Status)
	}
	return fmt.Sprintf("%s → %s, %s", e.ID, metricStr, e.Status)
}

func LearningOneLiner(l model.Learning) string {
	typ := string(l.Type)
	text := l.Text
	if len(text) > 80 {
		text = text[:77] + "..."
	}
	return fmt.Sprintf("[%s] %s", typ, text)
}

func GraveyardOneLiner(g model.GraveyardEntry) string {
	approach := g.Approach
	if len(approach) > 60 {
		approach = approach[:57] + "..."
	}
	reason := g.Reason
	if len(reason) > 60 {
		reason = reason[:57] + "..."
	}
	if g.ExperimentID != "" {
		return fmt.Sprintf("✗ %s — %s (%s)", approach, reason, g.ExperimentID)
	}
	return fmt.Sprintf("✗ %s — %s", approach, reason)
}

func ChangelogOneLiner(c model.ChangelogEntry) string {
	ts := c.Timestamp.Format("2006-01-02 15:04")
	if c.Summary != "" {
		return fmt.Sprintf("[%s] %s: %s", ts, c.Action, c.Summary)
	}
	if c.ID != "" {
		return fmt.Sprintf("[%s] %s: %s", ts, c.Action, c.ID)
	}
	return fmt.Sprintf("[%s] %s", ts, c.Action)
}
