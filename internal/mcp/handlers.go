package mcp

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rzzdr/marrow/internal/format"
	idx "github.com/rzzdr/marrow/internal/index"
	"github.com/rzzdr/marrow/internal/model"
	"github.com/rzzdr/marrow/internal/store"
	"github.com/rzzdr/marrow/internal/util"
)

type handlers struct {
	store *store.Store
	mu    sync.Mutex
}

// formatWarnings formats a slice of warnings into a user-friendly string
func formatWarnings(warnings []string) string {
	if len(warnings) == 0 {
		return ""
	}
	return "\n\n⚠ Warnings:\n  - " + strings.Join(warnings, "\n  - ")
}

func (h *handlers) getProjectSummary(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	proj, err := h.store.ReadProject()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read project: %v", err)), nil
	}

	index, err := h.store.ReadIndex()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read index: %v", err)), nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Project: %s\n", proj.Name)
	if proj.Description != "" {
		fmt.Fprintf(&b, "Description: %s\n", proj.Description)
	}
	fmt.Fprintf(&b, "Task: %s\nMetric: %s (%s)\n", proj.TaskType, proj.Metric.Name, proj.Metric.Direction)
	fmt.Fprintf(&b, "\n--- Index ---\n")
	c := index.Computed
	fmt.Fprintf(&b, "Experiments: %d\n", c.TotalExperiments)
	if c.BestExperiment != "" && c.BestMetric != nil {
		fmt.Fprintf(&b, "Best: %s (%s = %.4f)\n", c.BestExperiment, c.BestMetric.Name, c.BestMetric.Value)
	}
	if len(c.ExperimentChain) > 0 {
		fmt.Fprintf(&b, "Chain: %s\n", strings.Join(c.ExperimentChain, " → "))
	}
	fmt.Fprintf(&b, "Proven: %d | Assumptions: %d | Graveyard: %d\n", c.ProvenCount, c.AssumptionCount, c.GraveyardCount)

	p := index.Pinned
	if len(p.DoNotTry) > 0 {
		fmt.Fprintf(&b, "\nDo Not Try:\n")
		for _, d := range p.DoNotTry {
			fmt.Fprintf(&b, "  - %s\n", d)
		}
	}
	if len(p.DataWarnings) > 0 {
		fmt.Fprintf(&b, "\nData Warnings:\n")
		for _, w := range p.DataWarnings {
			fmt.Fprintf(&b, "  - %s\n", w)
		}
	}
	if len(p.Deferred) > 0 {
		fmt.Fprintf(&b, "\nDeferred:\n")
		for _, d := range p.Deferred {
			fmt.Fprintf(&b, "  - %s\n", d)
		}
	}
	if p.Notes != "" {
		fmt.Fprintf(&b, "\nNotes: %s\n", p.Notes)
	}

	text := b.String()
	return toolResultWithMeta(text, format.EstimateTokens(text), "summary"), nil
}

func (h *handlers) getBestExperiment(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	index, err := h.store.ReadIndex()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read index: %v", err)), nil
	}

	if index.Computed.BestExperiment == "" {
		return mcp.NewToolResultText("No experiments yet."), nil
	}

	exp, err := h.store.ReadExperiment(index.Computed.BestExperiment)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read experiment: %v", err)), nil
	}

	depth := model.ParseDepth(req.GetString("depth", "standard"))
	return experimentResult(exp, depth)
}

func (h *handlers) getExperiment(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: id"), nil
	}

	exp, err := h.store.ReadExperiment(id)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("experiment not found: %v", err)), nil
	}

	depth := model.ParseDepth(req.GetString("depth", "full"))
	return experimentResult(exp, depth)
}

func (h *handlers) getLearnings(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	lf, err := h.store.ReadLearnings()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read learnings: %v", err)), nil
	}

	typ := req.GetString("type", "all")
	depth := model.ParseDepth(req.GetString("depth", "summary"))

	var b strings.Builder

	if typ == "all" || typ == "proven" {
		if len(lf.Proven) > 0 {
			b.WriteString("Proven:\n")
			for _, l := range lf.Proven {
				fl := format.FilterLearning(l, depth)
				if depth == model.DepthSummary {
					fmt.Fprintf(&b, "  %s\n", format.LearningOneLiner(fl))
				} else {
					y, err := format.MarshalYAMLString(fl)
					if err != nil {
						return mcp.NewToolResultError(fmt.Sprintf("failed to marshal learning: %v", err)), nil
					}
					b.WriteString(y)
				}
			}
		}
	}

	if typ == "all" || typ == "assumption" {
		if len(lf.Assumptions) > 0 {
			b.WriteString("Assumptions:\n")
			for _, l := range lf.Assumptions {
				fl := format.FilterLearning(l, depth)
				if depth == model.DepthSummary {
					fmt.Fprintf(&b, "  %s\n", format.LearningOneLiner(fl))
				} else {
					y, err := format.MarshalYAMLString(fl)
					if err != nil {
						return mcp.NewToolResultError(fmt.Sprintf("failed to marshal learning: %v", err)), nil
					}
					b.WriteString(y)
				}
			}
		}
	}

	if b.Len() == 0 {
		return mcp.NewToolResultText("No learnings yet."), nil
	}

	text := b.String()
	return toolResultWithMeta(text, format.EstimateTokens(text), string(depth)), nil
}

func (h *handlers) getFailures(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	gf, err := h.store.ReadGraveyard()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read graveyard: %v", err)), nil
	}

	if len(gf.Entries) == 0 {
		return mcp.NewToolResultText("Graveyard is empty."), nil
	}

	depth := model.ParseDepth(req.GetString("depth", "summary"))

	var b strings.Builder
	b.WriteString("Failed Approaches:\n")
	for _, g := range gf.Entries {
		if depth == model.DepthSummary {
			fmt.Fprintf(&b, "  %s\n", format.GraveyardOneLiner(g))
		} else {
			y, err := format.MarshalYAMLString(g)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to marshal graveyard entry: %v", err)), nil
			}
			b.WriteString(y)
		}
	}

	text := b.String()
	return toolResultWithMeta(text, format.EstimateTokens(text), string(depth)), nil
}

func (h *handlers) getDataContext(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: name"), nil
	}

	raw, err := h.store.ReadContextRaw(name)
	if err != nil {
		names, _ := h.store.ListContextFiles()
		return mcp.NewToolResultError(
			fmt.Sprintf("context %q not found. Available: %s", name, strings.Join(names, ", ")),
		), nil
	}

	return toolResultWithMeta(raw, format.EstimateTokens(raw), "full"), nil
}

func (h *handlers) getChangelog(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sinceStr := req.GetString("since", "")

	var entries []model.ChangelogEntry
	if sinceStr != "" {
		t, err := time.Parse("2006-01-02", sinceStr)
		if err != nil {
			return mcp.NewToolResultError("invalid date format, use YYYY-MM-DD"), nil
		}
		entries, err = h.store.ReadChangelogSince(t)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to read changelog: %v", err)), nil
		}
	} else {
		cf, err := h.store.ReadChangelog()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to read changelog: %v", err)), nil
		}
		entries = cf.Entries
	}

	if len(entries) == 0 {
		return mcp.NewToolResultText("No changelog entries."), nil
	}

	var b strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&b, "%s\n", format.ChangelogOneLiner(e))
	}
	text := b.String()
	return toolResultWithMeta(text, format.EstimateTokens(text), "summary"), nil
}

func (h *handlers) getExperimentChain(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	index, err := h.store.ReadIndex()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read index: %v", err)), nil
	}

	if len(index.Computed.ExperimentChain) == 0 {
		return mcp.NewToolResultText("No experiment chain."), nil
	}

	depth := model.ParseDepth(req.GetString("depth", "summary"))

	var b strings.Builder
	for _, id := range index.Computed.ExperimentChain {
		exp, err := h.store.ReadExperiment(id)
		if err != nil {
			continue
		}
		if depth == model.DepthSummary {
			fmt.Fprintf(&b, "%s\n", format.ExperimentOneLiner(exp))
		} else {
			fe := format.FilterExperiment(exp, depth)
			y, err := format.MarshalYAMLString(fe)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to marshal experiment: %v", err)), nil
			}
			b.WriteString(y)
			b.WriteString("---\n")
		}
	}

	text := b.String()
	return toolResultWithMeta(text, format.EstimateTokens(text), string(depth)), nil
}

func (h *handlers) getExperimentsByTag(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tagsStr, err := req.RequireString("tags")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: tags"), nil
	}

	tags := util.SplitTags(tagsStr)

	exps, err := h.store.ListExperimentsByTag(tags)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list experiments: %v", err)), nil
	}

	if len(exps) == 0 {
		return mcp.NewToolResultText("No experiments match those tags."), nil
	}

	depth := model.ParseDepth(req.GetString("depth", "summary"))
	return experimentsResult(exps, depth)
}

func (h *handlers) compareExperiments(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id1, err := req.RequireString("id1")
	if err != nil {
		return mcp.NewToolResultError("missing id1"), nil
	}
	id2, err := req.RequireString("id2")
	if err != nil {
		return mcp.NewToolResultError("missing id2"), nil
	}

	exp1, err := h.store.ReadExperiment(id1)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("experiment %s not found", id1)), nil
	}
	exp2, err := h.store.ReadExperiment(id2)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("experiment %s not found", id2)), nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Comparison: %s vs %s\n\n", id1, id2)
	fmt.Fprintf(&b, "%s:\n  %s = %.4f | status: %s | model: %s\n",
		id1, exp1.Metric.Name, exp1.Metric.Value, exp1.Status, exp1.BaseModel)
	fmt.Fprintf(&b, "%s:\n  %s = %.4f | status: %s | model: %s\n",
		id2, exp2.Metric.Name, exp2.Metric.Value, exp2.Status, exp2.BaseModel)

	delta := exp2.Metric.Value - exp1.Metric.Value
	metricDirection := "higher_is_better" // default when project is unreadable
	var warnings []string
	proj, err := h.store.ReadProject()
	if err == nil {
		metricDirection = proj.Metric.Direction
	} else {
		warnings = append(warnings, fmt.Sprintf("project unreadable, assuming higher_is_better: %v", err))
	}
	direction := "improvement"
	if (strings.EqualFold(metricDirection, "higher_is_better") && delta < 0) ||
		(strings.EqualFold(metricDirection, "lower_is_better") && delta > 0) {
		direction = "regression"
	}
	if delta == 0 {
		direction = "no change"
	}
	fmt.Fprintf(&b, "\nDelta: %+.4f (%s)\n", delta, direction)

	if exp2.Notes != "" {
		fmt.Fprintf(&b, "\n%s notes: %s\n", id2, exp2.Notes)
	}

	text := b.String() + formatWarnings(warnings)
	return toolResultWithMeta(text, format.EstimateTokens(text), "standard"), nil
}

func (h *handlers) getAllExperiments(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	exps, err := h.store.ListExperiments()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list experiments: %v", err)), nil
	}

	if len(exps) == 0 {
		return mcp.NewToolResultText("No experiments yet."), nil
	}

	limit := int(req.GetFloat("limit", 0))
	if limit > 0 && len(exps) > limit {
		exps = exps[len(exps)-limit:]
	}

	depth := model.ParseDepth(req.GetString("depth", "summary"))
	return experimentsResult(exps, depth)
}

func (h *handlers) logExperiment(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	proj, err := h.store.ReadProject()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read project: %v", err)), nil
	}

	status, err := req.RequireString("status")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: status"), nil
	}
	validStatuses := map[string]bool{"improved": true, "degraded": true, "neutral": true, "failed": true}
	if !validStatuses[status] {
		return mcp.NewToolResultError(fmt.Sprintf("invalid status: %s. Use: improved|degraded|neutral|failed", status)), nil
	}

	id, err := h.store.NextExperimentID()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to generate ID: %v", err)), nil
	}

	metricVal := req.GetFloat("metric_value", 0)

	exp := model.Experiment{
		ID:        id,
		Timestamp: time.Now().UTC(),
		BaseModel: req.GetString("base_model", ""),
		Status:    status,
		Metric: model.MetricResult{
			Name:  proj.Metric.Name,
			Value: metricVal,
		},
		Notes: req.GetString("notes", ""),
	}

	parents := req.GetString("parents", "")
	if parents != "" {
		exp.Parents = util.SplitTags(parents)
		for _, pid := range exp.Parents {
			if _, err := h.store.ReadExperiment(pid); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("parent experiment %s not found", pid)), nil
			}
		}
	}
	tags := req.GetString("tags", "")
	if tags != "" {
		exp.Tags = util.SplitTags(tags)
	}

	var warnings []string

	// Compute delta relative to best parent or current best
	if len(exp.Parents) > 0 {
		if parent, err := h.store.ReadExperiment(exp.Parents[0]); err == nil {
			exp.Metric.Baseline = parent.Metric.Value
			exp.Metric.Delta = exp.Metric.Value - parent.Metric.Value
		} else {
			warnings = append(warnings, fmt.Sprintf("could not compute baseline from parent: %v", err))
		}
	} else {
		curIdx, err := h.store.ReadIndex()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("could not compute baseline from index: %v", err))
		} else {
			if curIdx.Computed.BestMetric != nil {
				exp.Metric.Baseline = curIdx.Computed.BestMetric.Value
				exp.Metric.Delta = exp.Metric.Value - curIdx.Computed.BestMetric.Value
			}
		}
	}

	if err := h.store.WriteExperiment(exp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to write experiment: %v", err)), nil
	}

	if _, err := idx.UpdateIncremental(h.store, exp); err != nil {
		warnings = append(warnings, fmt.Sprintf("index update failed: %v", err))
	}

	if err := h.store.AppendChangelog(model.ChangelogEntry{
		Action:  "exp_logged",
		ID:      id,
		Summary: format.ExperimentOneLiner(exp),
	}); err != nil {
		warnings = append(warnings, fmt.Sprintf("changelog append failed: %v", err))
	}

	result := fmt.Sprintf("Logged experiment %s (%s = %.4f, %s)", id, proj.Metric.Name, metricVal, status)
	result += formatWarnings(warnings)
	return mcp.NewToolResultText(result), nil
}

func (h *handlers) addLearning(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	text, err := req.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError("missing text"), nil
	}
	typ, err := req.RequireString("type")
	if err != nil {
		return mcp.NewToolResultError("missing type"), nil
	}
	if typ != "proven" && typ != "assumption" {
		return mcp.NewToolResultError(fmt.Sprintf("invalid type: %s. Use: proven|assumption", typ)), nil
	}

	l := model.Learning{
		Type: model.LearningType(typ),
		Text: text,
	}
	tags := req.GetString("tags", "")
	if tags != "" {
		l.Tags = util.SplitTags(tags)
	}

	learnings, _ := h.store.ReadLearnings()
	graveyard, _ := h.store.ReadGraveyard()
	conflicts := idx.DetectConflicts(l, learnings, graveyard)

	id, err := h.store.AddLearning(l)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to add learning: %v", err)), nil
	}

	var warnings []string
	if err := h.store.AppendChangelog(model.ChangelogEntry{
		Action:  "learning_added",
		ID:      id,
		Type:    typ,
		Summary: text,
	}); err != nil {
		warnings = append(warnings, fmt.Sprintf("changelog append failed: %v", err))
	}

	if err := idx.UpdateLearningCounts(h.store); err != nil {
		warnings = append(warnings, fmt.Sprintf("learning counts update failed: %v", err))
	}

	result := fmt.Sprintf("Added learning %s [%s]", id, typ)
	result += formatWarnings(warnings)
	if len(conflicts) > 0 {
		result += "\n\n⚠ Potential conflicts:"
		for _, c := range conflicts {
			result += fmt.Sprintf("\n  - Conflicts with %s: %s", c.ConflictsWith, c.ConflictingEntry)
		}
	}

	return mcp.NewToolResultText(result), nil
}

func (h *handlers) addGraveyardEntry(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	approach, err := req.RequireString("approach")
	if err != nil {
		return mcp.NewToolResultError("missing approach"), nil
	}
	reason, err := req.RequireString("reason")
	if err != nil {
		return mcp.NewToolResultError("missing reason"), nil
	}

	g := model.GraveyardEntry{
		Approach:     approach,
		Reason:       reason,
		ExperimentID: req.GetString("experiment_id", ""),
	}
	tags := req.GetString("tags", "")
	if tags != "" {
		g.Tags = util.SplitTags(tags)
	}

	id, err := h.store.AddGraveyardEntry(g)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to add entry: %v", err)), nil
	}

	var warnings []string
	if err := h.store.AppendChangelog(model.ChangelogEntry{
		Action:  "graveyard_added",
		ID:      id,
		Summary: approach + " — " + reason,
	}); err != nil {
		warnings = append(warnings, fmt.Sprintf("changelog append failed: %v", err))
	}

	if err := idx.UpdateLearningCounts(h.store); err != nil {
		warnings = append(warnings, fmt.Sprintf("learning counts update failed: %v", err))
	}

	result := fmt.Sprintf("Added graveyard entry %s", id)
	result += formatWarnings(warnings)
	return mcp.NewToolResultText(result), nil
}

func (h *handlers) updatePinned(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	field, err := req.RequireString("field")
	if err != nil {
		return mcp.NewToolResultError("missing field"), nil
	}
	action, err := req.RequireString("action")
	if err != nil {
		return mcp.NewToolResultError("missing action"), nil
	}
	value, err := req.RequireString("value")
	if err != nil {
		return mcp.NewToolResultError("missing value"), nil
	}

	index, err := h.store.ReadIndex()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read index: %v", err)), nil
	}

	p := &index.Pinned
	var target *[]string

	switch field {
	case "do_not_try":
		target = &p.DoNotTry
	case "deferred":
		target = &p.Deferred
	case "data_warnings":
		target = &p.DataWarnings
	case "critical_features":
		target = &p.CriticalFeatures
	case "notes":
		if action == "set" {
			p.Notes = value
		} else {
			p.Notes += "\n" + value
		}
		if err := h.store.WriteIndex(index); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to write index: %v", err)), nil
		}
		var warnings []string
		if err := h.store.AppendChangelog(model.ChangelogEntry{
			Action:  "pinned_updated",
			Summary: "notes updated",
		}); err != nil {
			warnings = append(warnings, fmt.Sprintf("changelog append failed: %v", err))
		}
		result := "Updated notes." + formatWarnings(warnings)
		return mcp.NewToolResultText(result), nil
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unknown field: %s. Use: do_not_try, deferred, data_warnings, critical_features, notes", field)), nil
	}

	switch action {
	case "add":
		exists := false
		for _, v := range *target {
			if v == value {
				exists = true
				break
			}
		}
		if !exists {
			*target = append(*target, value)
		}
	case "remove":
		filtered := (*target)[:0]
		for _, v := range *target {
			if v != value {
				filtered = append(filtered, v)
			}
		}
		*target = filtered
	case "set":
		*target = []string{value}
	default:
		return mcp.NewToolResultError(fmt.Sprintf("unknown action: %s. Use: add, remove, set", action)), nil
	}

	if err := h.store.WriteIndex(index); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to write index: %v", err)), nil
	}

	var warnings []string
	if err := h.store.AppendChangelog(model.ChangelogEntry{
		Action:  "pinned_updated",
		Summary: fmt.Sprintf("%s %s: %s", action, field, value),
	}); err != nil {
		warnings = append(warnings, fmt.Sprintf("changelog append failed: %v", err))
	}

	result := fmt.Sprintf("Updated pinned.%s (%s: %s)", field, action, value)
	result += formatWarnings(warnings)
	return mcp.NewToolResultText(result), nil
}

func (h *handlers) getPrelude(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	intent, err := req.RequireString("intent")
	if err != nil {
		return mcp.NewToolResultError("missing intent"), nil
	}

	intentLower := strings.ToLower(intent)

	var b strings.Builder

	proj, err := h.store.ReadProject()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read project: %v", err)), nil
	}
	// best-effort: index may not exist yet, fields default to zero values
	index, _ := h.store.ReadIndex()

	fmt.Fprintf(&b, "Project: %s | Task: %s | Metric: %s (%s)\n", proj.Name, proj.TaskType, proj.Metric.Name, proj.Metric.Direction)
	c := index.Computed
	if c.BestExperiment != "" && c.BestMetric != nil {
		fmt.Fprintf(&b, "Best: %s (%s = %.4f)\n", c.BestExperiment, c.BestMetric.Name, c.BestMetric.Value)
	}

	if containsAny(intentLower, "feature", "eda", "data", "column", "variable") {
		b.WriteString("\n--- Data Context ---\n")
		contextNames, _ := h.store.ListContextFiles()
		for _, name := range contextNames {
			nameLower := strings.ToLower(name)
			if containsAny(nameLower, "eda", "feature", "data", "column", "variable", "pipeline", "overview") ||
				strings.Contains(intentLower, nameLower) {
				if raw, err := h.store.ReadContextRaw(name); err == nil {
					fmt.Fprintf(&b, "[%s]\n%s\n", name, raw)
				}
			}
		}
		if len(index.Pinned.DataWarnings) > 0 {
			b.WriteString("Data Warnings:\n")
			for _, w := range index.Pinned.DataWarnings {
				fmt.Fprintf(&b, "  - %s\n", w)
			}
		}
	}

	if containsAny(intentLower, "hyperparameter", "tune", "tuning", "lr", "learning rate", "param") {
		b.WriteString("\n--- HP Tuning Context ---\n")
		exps, _ := h.store.ListExperimentsByTag([]string{"lr_tuning", "hp_tuning", "tuning", "hyperparameter"})
		for _, e := range exps {
			fmt.Fprintf(&b, "  %s\n", format.ExperimentOneLiner(e))
		}
	}

	if containsAny(intentLower, "fail", "error", "avoid", "not work", "graveyard", "wrong") {
		b.WriteString("\n--- Failures ---\n")
		gf, _ := h.store.ReadGraveyard()
		for _, g := range gf.Entries {
			fmt.Fprintf(&b, "  %s\n", format.GraveyardOneLiner(g))
		}
		if len(index.Pinned.DoNotTry) > 0 {
			b.WriteString("Do Not Try:\n")
			for _, d := range index.Pinned.DoNotTry {
				fmt.Fprintf(&b, "  - %s\n", d)
			}
		}
	}

	if containsAny(intentLower, "model", "architecture", "network", "backbone") {
		b.WriteString("\n--- Best Experiment (full) ---\n")
		if c.BestExperiment != "" {
			if exp, err := h.store.ReadExperiment(c.BestExperiment); err == nil {
				y, _ := format.MarshalYAMLString(exp)
				b.WriteString(y)
			}
		}
	}

	lf, _ := h.store.ReadLearnings()
	if len(lf.Proven) > 0 {
		b.WriteString("\n--- Proven Learnings ---\n")
		for _, l := range lf.Proven {
			fmt.Fprintf(&b, "  %s\n", format.LearningOneLiner(l))
		}
	}

	text := b.String()
	return toolResultWithMeta(text, format.EstimateTokens(text), "prelude"), nil
}

func experimentResult(exp model.Experiment, depth model.Depth) (*mcp.CallToolResult, error) {
	if depth == model.DepthSummary {
		text := format.ExperimentOneLiner(exp)
		return toolResultWithMeta(text, format.EstimateTokens(text), "summary"), nil
	}
	fe := format.FilterExperiment(exp, depth)
	y, err := format.MarshalYAMLString(fe)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("marshaling failed: %v", err)), nil
	}
	return toolResultWithMeta(y, format.EstimateTokens(y), string(depth)), nil
}

func experimentsResult(exps []model.Experiment, depth model.Depth) (*mcp.CallToolResult, error) {
	var b strings.Builder
	for _, e := range exps {
		if depth == model.DepthSummary {
			fmt.Fprintf(&b, "%s\n", format.ExperimentOneLiner(e))
		} else {
			fe := format.FilterExperiment(e, depth)
			y, err := format.MarshalYAMLString(fe)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to marshal experiment: %v", err)), nil
			}
			b.WriteString(y)
			b.WriteString("---\n")
		}
	}
	text := b.String()
	return toolResultWithMeta(text, format.EstimateTokens(text), string(depth)), nil
}

func toolResultWithMeta(text string, tokensApprox int, depth string) *mcp.CallToolResult {
	header := fmt.Sprintf("[tokens≈%d depth=%s]\n", tokensApprox, depth)
	return mcp.NewToolResultText(header + text)
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
