package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ctxCmd = &cobra.Command{
	Use:   "ctx",
	Short: "Manage freeform context files",
}

var ctxListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all context files",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := getStoreFromRoot()

		names, err := s.ListContextFiles()
		if err != nil {
			return err
		}

		if len(names) == 0 {
			fmt.Println("No context files yet.")
			fmt.Println("Add files to .marrow/context/ (e.g. eda.yaml, features.yaml)")
			return nil
		}

		for _, n := range names {
			fmt.Println("  " + n)
		}
		return nil
	},
}

var ctxShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show a context file's contents",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s := getStoreFromRoot()

		raw, err := s.ReadContextRaw(args[0])
		if err != nil {
			return fmt.Errorf("reading context %s: %w", args[0], err)
		}
		fmt.Print(raw)
		return nil
	},
}

func init() {
	ctxCmd.AddCommand(ctxListCmd)
	ctxCmd.AddCommand(ctxShowCmd)
}
