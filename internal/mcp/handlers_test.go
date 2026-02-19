package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/rzzdr/marrow/internal/model"
	"github.com/rzzdr/marrow/internal/store"
)

// helper to create a temp .marrow store for tests
func setupTestStore(t *testing.T) *store.Store {
	t.Helper()
	dir := t.TempDir()
	s := store.New(dir)
	proj := model.Project{
		Name: "test-project",
		Metric: model.MetricDef{
			Name:      "accuracy",
			Direction: "higher_is_better",
		},
		TaskType: "classification",
	}
	if err := s.Init(proj); err != nil {
		t.Fatalf("failed to init store: %v", err)
	}
	return s
}

// helper to create a CallToolRequest with given arguments
func makeRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

// helper to extract text from CallToolResult
func resultText(r *mcp.CallToolResult) string {
	if r == nil || len(r.Content) == 0 {
		return ""
	}
	if tc, ok := r.Content[0].(mcp.TextContent); ok {
		return tc.Text
	}
	return ""
}

func TestFormatWarnings(t *testing.T) {
	t.Run("empty warnings returns empty string", func(t *testing.T) {
		result := formatWarnings(nil)
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
		result = formatWarnings([]string{})
		if result != "" {
			t.Errorf("expected empty string, got %q", result)
		}
	})

	t.Run("single warning", func(t *testing.T) {
		result := formatWarnings([]string{"something failed"})
		if !strings.Contains(result, "⚠ Warnings:") {
			t.Errorf("expected warning header, got %q", result)
		}
		if !strings.Contains(result, "something failed") {
			t.Errorf("expected warning text, got %q", result)
		}
	})

	t.Run("multiple warnings", func(t *testing.T) {
		result := formatWarnings([]string{"warn1", "warn2"})
		if !strings.Contains(result, "warn1") || !strings.Contains(result, "warn2") {
			t.Errorf("expected both warnings, got %q", result)
		}
	})
}

func TestGetProjectSummary_StoreErrors(t *testing.T) {
	t.Run("missing project file returns error", func(t *testing.T) {
		dir := t.TempDir()
		s := store.New(dir)
		// Don't init — project file won't exist
		h := &handlers{store: s}

		result, err := h.getProjectSummary(context.Background(), makeRequest(nil))
		if err != nil {
			t.Fatalf("unexpected Go error: %v", err)
		}
		if !result.IsError {
			t.Error("expected tool error result when project is missing")
		}
		text := resultText(result)
		if !strings.Contains(text, "failed to read project") {
			t.Errorf("expected 'failed to read project' in error, got %q", text)
		}
	})
}

func TestLogExperiment_WarningsOnChangelogFailure(t *testing.T) {
	s := setupTestStore(t)
	h := &handlers{store: s}

	// Log first experiment normally
	result, err := h.logExperiment(context.Background(), makeRequest(map[string]any{
		"status":       "improved",
		"metric_value": 0.85,
		"base_model":   "xgboost",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", resultText(result))
	}
	text := resultText(result)
	if !strings.Contains(text, "Logged experiment") {
		t.Errorf("expected success message, got %q", text)
	}

	// Make changelog unwritable to trigger warning path
	changelogPath := filepath.Join(s.Root(), "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	result, err = h.logExperiment(context.Background(), makeRequest(map[string]any{
		"status":       "neutral",
		"metric_value": 0.80,
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	// Primary operation should still succeed
	if result.IsError {
		t.Fatalf("expected success with warnings, got error: %s", resultText(result))
	}
	text = resultText(result)
	if !strings.Contains(text, "Logged experiment") {
		t.Errorf("expected success message, got %q", text)
	}
	// But should contain warning about changelog
	if !strings.Contains(text, "⚠ Warnings:") {
		t.Errorf("expected warnings in result, got %q", text)
	}
	if !strings.Contains(text, "changelog") {
		t.Errorf("expected changelog warning, got %q", text)
	}
}

func TestLogExperiment_InvalidStatus(t *testing.T) {
	s := setupTestStore(t)
	h := &handlers{store: s}

	result, err := h.logExperiment(context.Background(), makeRequest(map[string]any{
		"status":       "bogus",
		"metric_value": 0.5,
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid status")
	}
	text := resultText(result)
	if !strings.Contains(text, "invalid status") {
		t.Errorf("expected 'invalid status' message, got %q", text)
	}
}

func TestAddLearning_WarningsOnChangelogFailure(t *testing.T) {
	s := setupTestStore(t)
	h := &handlers{store: s}

	// Make changelog unwritable
	changelogPath := filepath.Join(s.Root(), "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	result, err := h.addLearning(context.Background(), makeRequest(map[string]any{
		"text": "test learning",
		"type": "proven",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	// Primary operation should succeed
	if result.IsError {
		t.Fatalf("expected success with warnings, got error: %s", resultText(result))
	}
	text := resultText(result)
	if !strings.Contains(text, "Added learning") {
		t.Errorf("expected success message, got %q", text)
	}
	if !strings.Contains(text, "⚠ Warnings:") {
		t.Errorf("expected warnings in result, got %q", text)
	}
}

func TestAddGraveyardEntry_WarningsOnChangelogFailure(t *testing.T) {
	s := setupTestStore(t)
	h := &handlers{store: s}

	// Make changelog unwritable
	changelogPath := filepath.Join(s.Root(), "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	result, err := h.addGraveyardEntry(context.Background(), makeRequest(map[string]any{
		"approach": "tried approach X",
		"reason":   "did not improve metric",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success with warnings, got error: %s", resultText(result))
	}
	text := resultText(result)
	if !strings.Contains(text, "Added graveyard entry") {
		t.Errorf("expected success message, got %q", text)
	}
	if !strings.Contains(text, "⚠ Warnings:") {
		t.Errorf("expected warnings, got %q", text)
	}
}

func TestUpdatePinned_WarningsOnChangelogFailure(t *testing.T) {
	s := setupTestStore(t)
	h := &handlers{store: s}

	// Make changelog unwritable
	changelogPath := filepath.Join(s.Root(), "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	result, err := h.updatePinned(context.Background(), makeRequest(map[string]any{
		"field":  "do_not_try",
		"action": "add",
		"value":  "never try random forests",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success with warnings, got error: %s", resultText(result))
	}
	text := resultText(result)
	if !strings.Contains(text, "Updated pinned") {
		t.Errorf("expected success message, got %q", text)
	}
	if !strings.Contains(text, "⚠ Warnings:") {
		t.Errorf("expected warnings, got %q", text)
	}
}

func TestUpdatePinned_NotesChangelogWarning(t *testing.T) {
	s := setupTestStore(t)
	h := &handlers{store: s}

	// Make changelog unwritable
	changelogPath := filepath.Join(s.Root(), "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	result, err := h.updatePinned(context.Background(), makeRequest(map[string]any{
		"field":  "notes",
		"action": "set",
		"value":  "some note",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success with warnings, got error: %s", resultText(result))
	}
	text := resultText(result)
	if !strings.Contains(text, "Updated notes") {
		t.Errorf("expected 'Updated notes' message, got %q", text)
	}
	if !strings.Contains(text, "⚠ Warnings:") {
		t.Errorf("expected warnings for changelog failure, got %q", text)
	}
}

func TestGetExperiment_NotFound(t *testing.T) {
	s := setupTestStore(t)
	h := &handlers{store: s}

	result, err := h.getExperiment(context.Background(), makeRequest(map[string]any{
		"id": "exp_999",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for non-existent experiment")
	}
	text := resultText(result)
	if !strings.Contains(text, "experiment not found") {
		t.Errorf("expected 'experiment not found', got %q", text)
	}
}

func TestGetLearnings_EmptyStore(t *testing.T) {
	s := setupTestStore(t)
	h := &handlers{store: s}

	result, err := h.getLearnings(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(result))
	}
	text := resultText(result)
	if !strings.Contains(text, "No learnings yet") {
		t.Errorf("expected 'No learnings yet', got %q", text)
	}
}

func TestGetFailures_EmptyStore(t *testing.T) {
	s := setupTestStore(t)
	h := &handlers{store: s}

	result, err := h.getFailures(context.Background(), makeRequest(map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", resultText(result))
	}
	text := resultText(result)
	if !strings.Contains(text, "Graveyard is empty") {
		t.Errorf("expected 'Graveyard is empty', got %q", text)
	}
}

func TestLogExperiment_BaselineWarningOnCorruptIndex(t *testing.T) {
	s := setupTestStore(t)
	h := &handlers{store: s}

	// Corrupt the index file
	indexPath := filepath.Join(s.Root(), "index.yaml")
	if err := os.WriteFile(indexPath, []byte("not: valid: yaml: [[["), 0644); err != nil {
		t.Fatalf("failed to corrupt index: %v", err)
	}

	result, err := h.logExperiment(context.Background(), makeRequest(map[string]any{
		"status":       "improved",
		"metric_value": 0.85,
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	// Should still succeed with warnings
	if result.IsError {
		t.Fatalf("expected success with warnings, got error: %s", resultText(result))
	}
	text := resultText(result)
	if !strings.Contains(text, "Logged experiment") {
		t.Errorf("expected success message, got %q", text)
	}
	if !strings.Contains(text, "⚠ Warnings:") {
		t.Errorf("expected warnings about baseline/index, got %q", text)
	}
}
