package cli

import (
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

func getStore() *store.Store {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	return store.New(dir)
}

func findMarrowRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".marrow")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	cwd, _ := os.Getwd()
	return cwd
}

func getStoreFromRoot() *store.Store {
	return store.New(findMarrowRoot())
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
}
