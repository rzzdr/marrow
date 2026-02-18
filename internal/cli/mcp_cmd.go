package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server for AI agent integration",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := getStoreFromRoot()
		if !s.Exists() {
			return fmt.Errorf("no .marrow/ found. Run 'marrow init' first")
		}
		return runMCPServer(s)
	},
}
