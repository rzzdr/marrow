package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rzzdr/marrow/internal/model"
	"github.com/rzzdr/marrow/internal/store"
)

// setupCLITestStore creates a temp marrow project and changes cwd to it.
// Returns the store and a cleanup function that restores the original cwd.
func setupCLITestStore(t *testing.T) (*store.Store, func()) {
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

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	return s, func() {
		os.Chdir(origDir)
	}
}

func TestGetStoreFromRoot_NoMarrowDir(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(dir)

	_, err := getStoreFromRoot()
	if err == nil {
		t.Error("expected error when no .marrow/ exists, got nil")
	}
	if !strings.Contains(err.Error(), "no .marrow/ directory found") {
		t.Errorf("expected descriptive error, got %q", err.Error())
	}
}

func TestGetStoreFromRoot_FindsMarrow(t *testing.T) {
	s, cleanup := setupCLITestStore(t)
	defer cleanup()
	_ = s

	got, err := getStoreFromRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestExpNewCmd_ExitCodeZeroOnSuccess(t *testing.T) {
	_, cleanup := setupCLITestStore(t)
	defer cleanup()

	// Reset command state
	expStatus = "improved"
	expMetric = 0.85
	expBaseModel = "xgboost"
	expNotes = "test"
	expParents = ""
	expTags = ""

	cmd := expNewCmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("expected exit code 0 (nil error), got error: %v", err)
	}
}

func TestExpNewCmd_ExitCodeOneOnMissingMarrow(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	cmd := expNewCmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Error("expected error when .marrow/ missing, got nil")
	}
}

func TestExpNewCmd_WarningsToStderrOnChangelogFailure(t *testing.T) {
	s, cleanup := setupCLITestStore(t)
	defer cleanup()

	// Make changelog unwritable to trigger warning
	changelogPath := filepath.Join(s.Root(), "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	expStatus = "improved"
	expMetric = 0.90
	expBaseModel = ""
	expNotes = ""
	expParents = ""
	expTags = ""

	cmd := expNewCmd
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)
	stdoutBuf := new(bytes.Buffer)
	cmd.SetOut(stdoutBuf)

	err := cmd.RunE(cmd, []string{})
	// Primary operation should succeed (exit 0)
	if err != nil {
		t.Errorf("expected exit code 0 with warnings, got error: %v", err)
	}

	// Warnings should appear on stderr
	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "warning:") {
		t.Errorf("expected warning on stderr, got %q", stderr)
	}
}

func TestLearnAddCmd_WarningsToStderrOnChangelogFailure(t *testing.T) {
	s, cleanup := setupCLITestStore(t)
	defer cleanup()

	// Make changelog unwritable
	changelogPath := filepath.Join(s.Root(), "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	learnType = "proven"
	learnTags = ""

	cmd := learnAddCmd
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := cmd.RunE(cmd, []string{"test learning text"})
	// Primary operation should succeed
	if err != nil {
		t.Errorf("expected nil error (exit 0) with warnings, got: %v", err)
	}

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "warning:") {
		t.Errorf("expected warning on stderr, got %q", stderr)
	}
}

func TestLearnGraveyardAddCmd_WarningsToStderrOnChangelogFailure(t *testing.T) {
	s, cleanup := setupCLITestStore(t)
	defer cleanup()

	// Make changelog unwritable
	changelogPath := filepath.Join(s.Root(), "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	graveApproach = "tried X"
	graveReason = "didn't work"
	graveExpID = ""
	graveTags = ""

	cmd := learnGraveyardAddCmd
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("expected nil error with warnings, got: %v", err)
	}

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "warning:") {
		t.Errorf("expected warning on stderr, got %q", stderr)
	}
}

func TestExpDeleteCmd_WarningsToStderrOnChangelogFailure(t *testing.T) {
	s, cleanup := setupCLITestStore(t)
	defer cleanup()

	// Create an experiment first
	expStatus = "improved"
	expMetric = 0.85
	expBaseModel = ""
	expNotes = ""
	expParents = ""
	expTags = ""

	newCmd := expNewCmd
	newCmd.SetOut(new(bytes.Buffer))
	newCmd.SetErr(new(bytes.Buffer))
	if err := newCmd.RunE(newCmd, []string{}); err != nil {
		t.Fatalf("failed to create experiment: %v", err)
	}

	// Make changelog unwritable
	changelogPath := filepath.Join(s.Root(), "changelog.yaml")
	if err := os.Chmod(changelogPath, 0000); err != nil {
		t.Skipf("cannot change file permissions: %v", err)
	}
	defer os.Chmod(changelogPath, 0644)

	cmd := expDeleteCmd
	stderrBuf := new(bytes.Buffer)
	cmd.SetErr(stderrBuf)
	cmd.SetOut(new(bytes.Buffer))

	err := cmd.RunE(cmd, []string{"exp_001"})
	// Primary operation should succeed
	if err != nil {
		t.Errorf("expected nil error with warnings, got: %v", err)
	}

	stderr := stderrBuf.String()
	if !strings.Contains(stderr, "warning:") {
		t.Errorf("expected warning on stderr, got %q", stderr)
	}
}
