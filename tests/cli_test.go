package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rzzdr/marrow/internal/model"
	"github.com/rzzdr/marrow/internal/store"
)

// buildBinary builds the marrow CLI binary and returns the path.
func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "marrow")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/marrow")
	cmd.Dir = filepath.Join(getRootDir(t))
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}
	return bin
}

// getRootDir returns the project root directory.
func getRootDir(t *testing.T) string {
	t.Helper()
	// Walk up from current test dir to find go.mod
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (go.mod)")
		}
		dir = parent
	}
}

// setupCLIProject creates a temp marrow project directory.
func setupCLIProject(t *testing.T) string {
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
	return dir
}

func TestCLI_ExpNew_ExitCodeZeroOnSuccess(t *testing.T) {
	bin := buildBinary(t)
	dir := setupCLIProject(t)

	cmd := exec.Command(bin, "exp", "new", "--metric", "0.85", "--status", "improved")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("expected exit 0, got error: %v\nOutput: %s", err, out)
	}
	if !strings.Contains(string(out), "Created experiment") {
		t.Errorf("expected 'Created experiment', got %q", string(out))
	}
}

func TestCLI_ExpNew_ExitCodeOneOnMissingMarrow(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir() // No .marrow/ here

	cmd := exec.Command(bin, "exp", "new", "--metric", "0.85", "--status", "improved")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("expected exit 1 when .marrow/ missing, got exit 0")
	}
	if !strings.Contains(string(out), "no .marrow/ directory found") {
		t.Errorf("expected descriptive error, got %q", string(out))
	}
}

func TestCLI_ExpNew_WarningsToStderrOnChangelogFailure(t *testing.T) {
	bin := buildBinary(t)
	dir := setupCLIProject(t)

	// Make changelog unwritable
	changelogPath := filepath.Join(dir, ".marrow", "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	cmd := exec.Command(bin, "exp", "new", "--metric", "0.90", "--status", "improved")
	cmd.Dir = dir

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	// Primary operation should succeed (exit 0)
	if err != nil {
		t.Errorf("expected exit 0 with warnings, got error: %v", err)
	}

	// Warnings should appear on stderr
	if !strings.Contains(stderr.String(), "warning:") {
		t.Errorf("expected warning on stderr, got %q", stderr.String())
	}
}

func TestCLI_ExpList_ExitCodeOneOnMissingMarrow(t *testing.T) {
	bin := buildBinary(t)
	dir := t.TempDir()

	cmd := exec.Command(bin, "exp", "list")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("expected exit 1 when .marrow/ missing")
	}
	if !strings.Contains(string(out), "no .marrow/ directory found") {
		t.Errorf("expected descriptive error, got %q", string(out))
	}
}

func TestCLI_LearnAdd_WarningsToStderrOnChangelogFailure(t *testing.T) {
	bin := buildBinary(t)
	dir := setupCLIProject(t)

	// Make changelog unwritable
	changelogPath := filepath.Join(dir, ".marrow", "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	cmd := exec.Command(bin, "learn", "add", "--type", "proven", "test learning text")
	cmd.Dir = dir

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Errorf("expected exit 0 with warnings, got error: %v", err)
	}
	if !strings.Contains(stderr.String(), "warning:") {
		t.Errorf("expected warning on stderr, got %q", stderr.String())
	}
}

func TestCLI_LearnGraveyardAdd_WarningsToStderrOnChangelogFailure(t *testing.T) {
	bin := buildBinary(t)
	dir := setupCLIProject(t)

	// Make changelog unwritable
	changelogPath := filepath.Join(dir, ".marrow", "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	cmd := exec.Command(bin, "learn", "graveyard", "--approach", "tried X", "--reason", "did not work")
	cmd.Dir = dir

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Errorf("expected exit 0 with warnings, got error: %v", err)
	}
	if !strings.Contains(stderr.String(), "warning:") {
		t.Errorf("expected warning on stderr, got %q", stderr.String())
	}
}

func TestCLI_ExpDelete_WarningsToStderrOnChangelogFailure(t *testing.T) {
	bin := buildBinary(t)
	dir := setupCLIProject(t)

	// Create an experiment first
	createCmd := exec.Command(bin, "exp", "new", "--metric", "0.85", "--status", "improved")
	createCmd.Dir = dir
	if out, err := createCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create experiment: %v\n%s", err, out)
	}

	// Make changelog unwritable
	changelogPath := filepath.Join(dir, ".marrow", "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	cmd := exec.Command(bin, "exp", "delete", "exp_001")
	cmd.Dir = dir

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Errorf("expected exit 0 with warnings, got error: %v", err)
	}
	if !strings.Contains(stderr.String(), "warning:") {
		t.Errorf("expected warning on stderr, got %q", stderr.String())
	}
}
