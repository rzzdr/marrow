package tests

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rzzdr/marrow/internal/mcp"
	"github.com/rzzdr/marrow/internal/model"
	"github.com/rzzdr/marrow/internal/store"
)

// setupTestStore creates a temp .marrow store for tests.
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

// callTool sends a tools/call JSON-RPC message to the server and returns the result.
func callTool(t *testing.T, srv *server.MCPServer, name string, args map[string]any) *gomcp.CallToolResult {
	t.Helper()
	req := gomcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      gomcp.NewRequestId(int64(1)),
		Request: gomcp.Request{
			Method: "tools/call",
		},
	}
	// Build raw message with params embedded
	raw := map[string]any{
		"jsonrpc": req.JSONRPC,
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": args,
		},
	}
	msg, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	resp := srv.HandleMessage(context.Background(), msg)
	if resp == nil {
		t.Fatal("nil response from HandleMessage")
	}

	jsonResp, ok := resp.(gomcp.JSONRPCResponse)
	if !ok {
		// Check if it's an error response
		jsonErr, errOk := resp.(gomcp.JSONRPCError)
		if errOk {
			t.Fatalf("JSON-RPC error: %v", jsonErr.Error)
		}
		t.Fatalf("unexpected response type: %T", resp)
	}

	result, ok := jsonResp.Result.(*gomcp.CallToolResult)
	if !ok {
		t.Fatalf("unexpected result type: %T", jsonResp.Result)
	}
	return result
}

// resultText extracts text from a CallToolResult.
func resultText(r *gomcp.CallToolResult) string {
	if r == nil || len(r.Content) == 0 {
		return ""
	}
	if tc, ok := r.Content[0].(gomcp.TextContent); ok {
		return tc.Text
	}
	return ""
}

func TestGetProjectSummary_MissingProject(t *testing.T) {
	dir := t.TempDir()
	s := store.New(dir)
	// Don't init — project file won't exist
	srv := mcp.NewServer(s)

	result := callTool(t, srv, "get_project_summary", nil)
	if !result.IsError {
		t.Error("expected tool error result when project is missing")
	}
	text := resultText(result)
	if !strings.Contains(text, "failed to read project") {
		t.Errorf("expected 'failed to read project' in error, got %q", text)
	}
}

func TestLogExperiment_WarningsOnChangelogFailure(t *testing.T) {
	s := setupTestStore(t)
	srv := mcp.NewServer(s)

	// Log first experiment normally
	result := callTool(t, srv, "log_experiment", map[string]any{
		"status":       "improved",
		"metric_value": 0.85,
		"base_model":   "xgboost",
	})
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

	result = callTool(t, srv, "log_experiment", map[string]any{
		"status":       "neutral",
		"metric_value": 0.80,
	})
	// Primary operation should still succeed
	if result.IsError {
		t.Fatalf("expected success with warnings, got error: %s", resultText(result))
	}
	text = resultText(result)
	if !strings.Contains(text, "Logged experiment") {
		t.Errorf("expected success message, got %q", text)
	}
	if !strings.Contains(text, "⚠ Warnings:") {
		t.Errorf("expected warnings in result, got %q", text)
	}
	if !strings.Contains(text, "changelog") {
		t.Errorf("expected changelog warning, got %q", text)
	}
}

func TestLogExperiment_InvalidStatus(t *testing.T) {
	s := setupTestStore(t)
	srv := mcp.NewServer(s)

	result := callTool(t, srv, "log_experiment", map[string]any{
		"status":       "bogus",
		"metric_value": 0.5,
	})
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
	srv := mcp.NewServer(s)

	// Make changelog unwritable
	changelogPath := filepath.Join(s.Root(), "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	result := callTool(t, srv, "add_learning", map[string]any{
		"text": "test learning",
		"type": "proven",
	})
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
	srv := mcp.NewServer(s)

	// Make changelog unwritable
	changelogPath := filepath.Join(s.Root(), "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	result := callTool(t, srv, "add_graveyard_entry", map[string]any{
		"approach": "tried approach X",
		"reason":   "did not improve metric",
	})
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
	srv := mcp.NewServer(s)

	// Make changelog unwritable
	changelogPath := filepath.Join(s.Root(), "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	result := callTool(t, srv, "update_pinned", map[string]any{
		"field":  "do_not_try",
		"action": "add",
		"value":  "never try random forests",
	})
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
	srv := mcp.NewServer(s)

	// Make changelog unwritable
	changelogPath := filepath.Join(s.Root(), "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	result := callTool(t, srv, "update_pinned", map[string]any{
		"field":  "notes",
		"action": "set",
		"value":  "some note",
	})
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
	srv := mcp.NewServer(s)

	result := callTool(t, srv, "get_experiment", map[string]any{
		"id": "exp_999",
	})
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
	srv := mcp.NewServer(s)

	result := callTool(t, srv, "get_learnings", map[string]any{})
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
	srv := mcp.NewServer(s)

	result := callTool(t, srv, "get_failures", map[string]any{})
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
	srv := mcp.NewServer(s)

	// Corrupt the index file
	indexPath := filepath.Join(s.Root(), "index.yaml")
	if err := os.WriteFile(indexPath, []byte("not: valid: yaml: [[["), 0644); err != nil {
		t.Fatalf("failed to corrupt index: %v", err)
	}

	result := callTool(t, srv, "log_experiment", map[string]any{
		"status":       "improved",
		"metric_value": 0.85,
	})
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

func TestCompareExperiments_WarningOnUnreadableProject(t *testing.T) {
	s := setupTestStore(t)
	srv := mcp.NewServer(s)

	// Log two experiments so we can compare them
	callTool(t, srv, "log_experiment", map[string]any{
		"status":       "improved",
		"metric_value": 0.80,
		"base_model":   "xgboost",
	})
	callTool(t, srv, "log_experiment", map[string]any{
		"status":       "improved",
		"metric_value": 0.90,
		"base_model":   "xgboost",
	})

	// Corrupt the project file so ReadProject fails
	projPath := filepath.Join(s.Root(), "marrow.yaml")
	if err := os.WriteFile(projPath, []byte("not: valid: yaml: [[["), 0644); err != nil {
		t.Fatalf("failed to corrupt project: %v", err)
	}

	// List experiments to get their IDs
	exps, err := s.ListExperiments()
	if err != nil || len(exps) < 2 {
		t.Fatalf("expected at least 2 experiments, got %d (err: %v)", len(exps), err)
	}

	result := callTool(t, srv, "compare_experiments", map[string]any{
		"id1": exps[0].ID,
		"id2": exps[1].ID,
	})
	if result.IsError {
		t.Fatalf("expected success with warning, got error: %s", resultText(result))
	}
	text := resultText(result)
	if !strings.Contains(text, "Comparison:") {
		t.Errorf("expected comparison output, got %q", text)
	}
	if !strings.Contains(text, "⚠ Warnings:") {
		t.Errorf("expected warnings about unreadable project, got %q", text)
	}
	if !strings.Contains(text, "assuming higher_is_better") {
		t.Errorf("expected fallback direction warning, got %q", text)
	}
}
