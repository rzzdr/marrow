package cli

import (
	"fmt"

	"github.com/rzzdr/marrow/internal/index"
	"github.com/rzzdr/marrow/internal/model"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Manage the index",
}

var indexRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Fully rebuild the index from experiment data",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}

		idx, err := index.Rebuild(s)
		if err != nil {
			return fmt.Errorf("rebuilding index: %w", err)
		}

		if err := s.AppendChangelog(model.ChangelogEntry{
			Action:  "index_rebuilt",
			Summary: fmt.Sprintf("%d experiments indexed", idx.Computed.TotalExperiments),
		}); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: failed to append changelog: %v\n", err)
		}

		fmt.Printf("Index rebuilt: %d experiments, best: %s\n",
			idx.Computed.TotalExperiments,
			idx.Computed.BestExperiment)
		return nil
	},
}

var indexShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current index",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}

		idx, err := s.ReadIndex()
		if err != nil {
			return fmt.Errorf("reading index: %w", err)
		}
		printIndex(idx)
		return nil
	},
}

func printIndex(idx model.Index) {
	c := idx.Computed
	fmt.Printf("Last updated:      %s\n", c.LastUpdated.Format("2006-01-02 15:04:05"))
	fmt.Printf("Total experiments:  %d\n", c.TotalExperiments)
	if c.BestExperiment != "" {
		fmt.Printf("Best experiment:   %s\n", c.BestExperiment)
	}
	if c.BestMetric != nil {
		fmt.Printf("Best metric:       %s = %.4f\n", c.BestMetric.Name, c.BestMetric.Value)
	}
	if len(c.ExperimentChain) > 0 {
		fmt.Printf("Experiment chain:  %v\n", c.ExperimentChain)
	}
	fmt.Printf("Proven learnings:  %d\n", c.ProvenCount)
	fmt.Printf("Assumptions:       %d\n", c.AssumptionCount)
	fmt.Printf("Graveyard entries: %d\n", c.GraveyardCount)

	p := idx.Pinned
	if len(p.DoNotTry) > 0 {
		fmt.Println("\n── Do Not Try ──")
		for _, d := range p.DoNotTry {
			fmt.Println("  ✗ " + d)
		}
	}
	if len(p.DataWarnings) > 0 {
		fmt.Println("\n── Data Warnings ──")
		for _, w := range p.DataWarnings {
			fmt.Println("  ⚠ " + w)
		}
	}
}

func init() {
	indexCmd.AddCommand(indexRebuildCmd)
	indexCmd.AddCommand(indexShowCmd)
}
