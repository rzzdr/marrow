package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rzzdr/marrow/internal/store"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "marrow",
	Short: "Structured knowledge base for AI research experiments",
	Long: `Marrow is a CLI tool and MCP server for storing, querying, and managing
structured context about AI research experiments, learnings, and data insights.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func findMarrowRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".marrow")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no .marrow/ directory found in any parent; run 'marrow init' first")
}

func getStoreFromRoot() (*store.Store, error) {
	root, err := findMarrowRoot()
	if err != nil {
		return nil, err
	}
	return store.New(root), nil
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(expCmd)
	rootCmd.AddCommand(learnCmd)
	rootCmd.AddCommand(ctxCmd)
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(summaryCmd)
	rootCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(versionCmd)
}
