package cli

import (
	"fmt"

	"github.com/rzzdr/marrow/internal/index"
	"github.com/rzzdr/marrow/internal/model"
	"github.com/rzzdr/marrow/internal/util"
	"github.com/spf13/cobra"
)

var learnCmd = &cobra.Command{
	Use:   "learn",
	Short: "Manage learnings and failed approaches",
}

var (
	learnType string
	learnTags string
)

var learnAddCmd = &cobra.Command{
	Use:   "add [text]",
	Short: "Add a learning",
	Long:  "Add a proven finding, assumption, or graveyard entry.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}
		if !s.Exists() {
			return fmt.Errorf("no .marrow/ found. Run 'marrow init' first")
		}

		if learnType != "proven" && learnType != "assumption" {
			return fmt.Errorf("invalid type %q: must be proven|assumption", learnType)
		}

		l := model.Learning{
			Type: model.LearningType(learnType),
			Text: args[0],
		}
		if learnTags != "" {
			l.Tags = util.SplitTags(learnTags)
		}

		learnings, _ := s.ReadLearnings()
		graveyard, _ := s.ReadGraveyard()
		conflicts := index.DetectConflicts(l, learnings, graveyard)
		if len(conflicts) > 0 {
			fmt.Println("⚠ Potential conflicts detected:")
			for _, c := range conflicts {
				fmt.Printf("  - Conflicts with %s: %s\n", c.ConflictsWith, c.ConflictingEntry)
			}
			fmt.Println("  (Adding anyway. Review and resolve manually.)")
		}

		id, err := s.AddLearning(l)
		if err != nil {
			return err
		}

		_ = index.UpdateLearningCounts(s)

		_ = s.AppendChangelog(model.ChangelogEntry{
			Action:  "learning_added",
			ID:      id,
			Type:    learnType,
			Summary: l.Text,
		})

		fmt.Printf("Added learning %s [%s]\n", id, learnType)
		return nil
	},
}

var learnListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all learnings",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}

		lf, err := s.ReadLearnings()
		if err != nil {
			return err
		}

		if len(lf.Proven) > 0 {
			fmt.Println("── Proven ──")
			for _, l := range lf.Proven {
				fmt.Printf("  %s: %s\n", l.ID, l.Text)
			}
		}
		if len(lf.Assumptions) > 0 {
			fmt.Println("── Assumptions ──")
			for _, l := range lf.Assumptions {
				fmt.Printf("  %s: %s\n", l.ID, l.Text)
			}
		}
		if len(lf.Proven) == 0 && len(lf.Assumptions) == 0 {
			fmt.Println("No learnings yet.")
		}
		return nil
	},
}

var (
	graveApproach string
	graveReason   string
	graveExpID    string
	graveTags     string
)

var learnGraveyardAddCmd = &cobra.Command{
	Use:   "graveyard",
	Short: "Add a failed approach to the graveyard",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}

		g := model.GraveyardEntry{
			Approach:     graveApproach,
			Reason:       graveReason,
			ExperimentID: graveExpID,
		}
		if graveTags != "" {
			g.Tags = util.SplitTags(graveTags)
		}

		id, err := s.AddGraveyardEntry(g)
		if err != nil {
			return err
		}

		_ = index.UpdateLearningCounts(s)

		_ = s.AppendChangelog(model.ChangelogEntry{
			Action:  "graveyard_added",
			ID:      id,
			Summary: g.Approach + " — " + g.Reason,
		})

		fmt.Printf("Added graveyard entry %s\n", id)
		return nil
	},
}

var learnGraveyardListCmd = &cobra.Command{
	Use:   "graveyard-list",
	Short: "List all failed approaches",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}

		gf, err := s.ReadGraveyard()
		if err != nil {
			return err
		}

		if len(gf.Entries) == 0 {
			fmt.Println("Graveyard is empty.")
			return nil
		}

		fmt.Println("── Graveyard ──")
		for _, g := range gf.Entries {
			fmt.Printf("  %s: %s — %s", g.ID, g.Approach, g.Reason)
			if g.ExperimentID != "" {
				fmt.Printf(" (%s)", g.ExperimentID)
			}
			fmt.Println()
		}
		return nil
	},
}

func init() {
	learnAddCmd.Flags().StringVar(&learnType, "type", "assumption", "Learning type: proven|assumption")
	learnAddCmd.Flags().StringVar(&learnTags, "tags", "", "Comma-separated tags")

	learnGraveyardAddCmd.Flags().StringVar(&graveApproach, "approach", "", "The approach that failed (required)")
	learnGraveyardAddCmd.Flags().StringVar(&graveReason, "reason", "", "Why it failed (required)")
	learnGraveyardAddCmd.Flags().StringVar(&graveExpID, "exp", "", "Related experiment ID")
	learnGraveyardAddCmd.Flags().StringVar(&graveTags, "tags", "", "Comma-separated tags")
	_ = learnGraveyardAddCmd.MarkFlagRequired("approach")
	_ = learnGraveyardAddCmd.MarkFlagRequired("reason")

	learnCmd.AddCommand(learnAddCmd)
	learnCmd.AddCommand(learnListCmd)
	learnCmd.AddCommand(learnGraveyardAddCmd)
	learnCmd.AddCommand(learnGraveyardListCmd)
	learnCmd.AddCommand(learnDeleteCmd)
	learnCmd.AddCommand(learnGraveyardDeleteCmd)
}

var learnDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a learning",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}

		if err := s.DeleteLearning(args[0]); err != nil {
			return err
		}

		_ = index.UpdateLearningCounts(s)

		_ = s.AppendChangelog(model.ChangelogEntry{
			Action:  "learning_deleted",
			ID:      args[0],
			Summary: "deleted learning " + args[0],
		})

		fmt.Printf("Deleted learning %s\n", args[0])
		return nil
	},
}

var learnGraveyardDeleteCmd = &cobra.Command{
	Use:   "graveyard-delete [id]",
	Short: "Delete a graveyard entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStoreFromRoot()
		if err != nil {
			return err
		}

		if err := s.DeleteGraveyardEntry(args[0]); err != nil {
			return err
		}

		_ = index.UpdateLearningCounts(s)

		_ = s.AppendChangelog(model.ChangelogEntry{
			Action:  "graveyard_deleted",
			ID:      args[0],
			Summary: "deleted graveyard entry " + args[0],
		})

		fmt.Printf("Deleted graveyard entry %s\n", args[0])
		return nil
	},
}
