package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Print a concise project summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}

		proj, err := s.ReadProject()
		if err != nil {
			return fmt.Errorf("reading project: %w", err)
		}

		idx, err := s.ReadIndex()
		if err != nil {
			return fmt.Errorf("reading index: %w", err)
		}

		fmt.Printf("Project: %s\n", proj.Name)
		if proj.Description != "" {
			fmt.Printf("  %s\n", proj.Description)
		}
		fmt.Printf("Task:    %s\n", proj.TaskType)
		fmt.Printf("Metric:  %s (%s)\n", proj.Metric.Name, proj.Metric.Direction)
		fmt.Println()
		printIndex(idx)
		return nil
	},
}
