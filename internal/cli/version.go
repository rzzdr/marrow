package cli

import (
	"fmt"

	"github.com/rzzdr/marrow/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("marrow %s (commit: %s, built: %s)\n", version.Version, version.Commit, version.Date)
	},
}
