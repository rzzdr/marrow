package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/rzzdr/marrow/internal/format"
	"github.com/rzzdr/marrow/internal/index"
	"github.com/rzzdr/marrow/internal/model"
	"github.com/spf13/cobra"
)

var expCmd = &cobra.Command{
	Use:   "exp",
	Short: "Manage experiments",
}

var (
	expBaseModel string
	expParents   string
	expMetric    float64
	expStatus    string
	expTags      string
	expNotes     string
)

var expNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new experiment record",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := getStoreFromRoot()
		if !s.Exists() {
			return fmt.Errorf("no .marrow/ found. Run 'marrow init' first")
		}

		proj, err := s.ReadProject()
		if err != nil {
			return err
		}

		id, err := s.NextExperimentID()
		if err != nil {
			return err
		}

		exp := model.Experiment{
			ID:        id,
			Timestamp: time.Now().UTC(),
			BaseModel: expBaseModel,
			Status:    expStatus,
			Metric: model.MetricResult{
				Name:  proj.Metric.Name,
				Value: expMetric,
			},
			Notes: expNotes,
		}

		if expParents != "" {
			exp.Parents = strings.Split(expParents, ",")
		}
		if expTags != "" {
			exp.Tags = strings.Split(expTags, ",")
		}

		if err := s.WriteExperiment(exp); err != nil {
			return err
		}

		if _, err := index.UpdateIncremental(s, exp); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: index update failed: %v\n", err)
		}

		_ = s.AppendChangelog(model.ChangelogEntry{
			Action:  "exp_logged",
			ID:      id,
			Summary: format.ExperimentOneLiner(exp),
		})

		fmt.Printf("Created experiment %s\n", id)
		return nil
	},
}

var expListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all experiments",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := getStoreFromRoot()

		exps, err := s.ListExperiments()
		if err != nil {
			return err
		}

		if len(exps) == 0 {
			fmt.Println("No experiments yet.")
			return nil
		}

		for _, e := range exps {
			fmt.Println(format.ExperimentOneLiner(e))
		}
		return nil
	},
}

var expShowCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show full details of an experiment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s := getStoreFromRoot()

		exp, err := s.ReadExperiment(args[0])
		if err != nil {
			return fmt.Errorf("reading experiment %s: %w", args[0], err)
		}

		data, err := format.MarshalYAMLString(exp)
		if err != nil {
			return err
		}
		fmt.Print(data)
		return nil
	},
}

func init() {
	expNewCmd.Flags().StringVar(&expBaseModel, "model", "", "Base model family (e.g. xgboost, resnet)")
	expNewCmd.Flags().StringVar(&expParents, "parents", "", "Comma-separated parent experiment IDs")
	expNewCmd.Flags().Float64Var(&expMetric, "metric", 0, "Primary metric value")
	expNewCmd.Flags().StringVar(&expStatus, "status", "neutral", "Outcome: improved|degraded|neutral|failed")
	expNewCmd.Flags().StringVar(&expTags, "tags", "", "Comma-separated tags")
	expNewCmd.Flags().StringVar(&expNotes, "notes", "", "Freeform notes")

	expCmd.AddCommand(expNewCmd)
	expCmd.AddCommand(expListCmd)
	expCmd.AddCommand(expShowCmd)
}
