package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/rzzdr/marrow/internal/format"
	"github.com/rzzdr/marrow/internal/index"
	"github.com/rzzdr/marrow/internal/model"
	"github.com/rzzdr/marrow/internal/util"
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

var validStatuses = map[string]bool{
	"improved": true, "degraded": true, "neutral": true, "failed": true,
}

var expNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new experiment record",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}
		if !s.Exists() {
			return fmt.Errorf("no .marrow/ found. Run 'marrow init' first")
		}

		if !validStatuses[expStatus] {
			return fmt.Errorf("invalid status %q: must be improved|degraded|neutral|failed", expStatus)
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
			exp.Parents = util.SplitTags(expParents)
			for _, pid := range exp.Parents {
				if _, err := s.ReadExperiment(pid); err != nil {
					return fmt.Errorf("parent experiment %q not found", pid)
				}
			}
		}
		if expTags != "" {
			exp.Tags = util.SplitTags(expTags)
		}

		// Compute delta relative to best parent or current best
		if len(exp.Parents) > 0 {
			if parent, err := s.ReadExperiment(exp.Parents[0]); err == nil {
				exp.Metric.Baseline = parent.Metric.Value
				exp.Metric.Delta = exp.Metric.Value - parent.Metric.Value
			}
		} else {
			idx, _ := s.ReadIndex()
			if idx.Computed.BestMetric != nil {
				exp.Metric.Baseline = idx.Computed.BestMetric.Value
				exp.Metric.Delta = exp.Metric.Value - idx.Computed.BestMetric.Value
			}
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

var (
	expListStatus string
	expListTag    string
	expListLimit  int
)

var (
	expEditNotes  string
	expEditStatus string
	expEditTags   string
)

var expListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all experiments",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}

		exps, err := s.ListExperiments()
		if err != nil {
			return err
		}

		if expListStatus != "" {
			var filtered []model.Experiment
			for _, e := range exps {
				if e.Status == expListStatus {
					filtered = append(filtered, e)
				}
			}
			exps = filtered
		}

		if expListTag != "" {
			wantTags := util.SplitTags(expListTag)
			wantSet := make(map[string]bool, len(wantTags))
			for _, t := range wantTags {
				wantSet[t] = true
			}
			var filtered []model.Experiment
			for _, e := range exps {
				for _, t := range e.Tags {
					if wantSet[t] {
						filtered = append(filtered, e)
						break
					}
				}
			}
			exps = filtered
		}

		if len(exps) == 0 {
			fmt.Println("No experiments match.")
			return nil
		}

		if expListLimit > 0 && len(exps) > expListLimit {
			exps = exps[len(exps)-expListLimit:]
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
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}

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
	_ = expNewCmd.MarkFlagRequired("metric")

	expListCmd.Flags().StringVar(&expListStatus, "status", "", "Filter by status: improved|degraded|neutral|failed")
	expListCmd.Flags().StringVar(&expListTag, "tag", "", "Filter by tag (comma-separated)")
	expListCmd.Flags().IntVar(&expListLimit, "limit", 0, "Show only the last N experiments")

	expEditCmd.Flags().StringVar(&expEditNotes, "notes", "", "New notes")
	expEditCmd.Flags().StringVar(&expEditStatus, "status", "", "New status: improved|degraded|neutral|failed")
	expEditCmd.Flags().StringVar(&expEditTags, "tags", "", "New comma-separated tags")

	expCmd.AddCommand(expNewCmd)
	expCmd.AddCommand(expListCmd)
	expCmd.AddCommand(expShowCmd)
	expCmd.AddCommand(expEditCmd)
	expCmd.AddCommand(expDeleteCmd)
}

var expEditCmd = &cobra.Command{
	Use:   "edit [id]",
	Short: "Edit an experiment's notes, status, or tags",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}

		exp, err := s.ReadExperiment(args[0])
		if err != nil {
			return fmt.Errorf("reading experiment %s: %w", args[0], err)
		}

		changed := false
		if cmd.Flags().Changed("notes") {
			exp.Notes = expEditNotes
			changed = true
		}
		if cmd.Flags().Changed("status") {
			if !validStatuses[expEditStatus] {
				return fmt.Errorf("invalid status %q: must be improved|degraded|neutral|failed", expEditStatus)
			}
			exp.Status = expEditStatus
			changed = true
		}
		if cmd.Flags().Changed("tags") {
			exp.Tags = util.SplitTags(expEditTags)
			changed = true
		}

		if !changed {
			return fmt.Errorf("nothing to edit; use --notes, --status, or --tags")
		}

		if err := s.WriteExperiment(exp); err != nil {
			return err
		}

		if _, err := index.Rebuild(s); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: index rebuild failed: %v\n", err)
		}

		_ = s.AppendChangelog(model.ChangelogEntry{
			Action:  "exp_edited",
			ID:      args[0],
			Summary: "edited experiment " + args[0],
		})

		fmt.Printf("Updated experiment %s\n", args[0])
		return nil
	},
}

var expDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete an experiment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}

		refs, err := s.FindParentRefs(args[0])
		if err != nil {
			return err
		}
		if len(refs) > 0 {
			return fmt.Errorf("cannot delete %s: referenced as parent by %s", args[0], strings.Join(refs, ", "))
		}

		if err := s.DeleteExperiment(args[0]); err != nil {
			return err
		}

		_ = s.AppendChangelog(model.ChangelogEntry{
			Action:  "exp_deleted",
			ID:      args[0],
			Summary: "deleted experiment " + args[0],
		})

		if _, err := index.Rebuild(s); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: index rebuild failed: %v\n", err)
		}

		fmt.Printf("Deleted experiment %s\n", args[0])
		return nil
	},
}
