package cli

import (
	"fmt"
	"os"

	"github.com/rzzdr/marrow/internal/model"
	"github.com/spf13/cobra"
)

var initTemplate string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Marrow project",
	Long:  "Scaffold the .marrow/ directory with project configuration.",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := getStore()
		if s.Exists() {
			return fmt.Errorf(".marrow/ already exists in this directory")
		}

		// Warn if a parent directory already has .marrow
		root := findMarrowRoot()
		cwd, _ := os.Getwd()
		if root != cwd {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: parent directory %s already has .marrow/; this will create a nested project\n", root)
		}

		project := model.Project{
			Name: "untitled",
			Metric: model.MetricDef{
				Name:      "metric",
				Direction: "higher_is_better",
			},
		}

		validTemplates := map[string]bool{
			"kaggle-tabular": true, "llm-finetune": true,
			"paper-replication": true, "rl-experiment": true,
		}
		if initTemplate != "" && !validTemplates[initTemplate] {
			return fmt.Errorf("unknown template %q: valid templates are kaggle-tabular, llm-finetune, paper-replication, rl-experiment", initTemplate)
		}

		switch initTemplate {
		case "kaggle-tabular":
			project.TaskType = "classification"
			project.Metric = model.MetricDef{
				Name:      "AUC-ROC",
				Direction: "higher_is_better",
			}
		case "llm-finetune":
			project.TaskType = "generation"
			project.Metric = model.MetricDef{
				Name:      "eval_loss",
				Direction: "lower_is_better",
			}
		case "paper-replication":
			project.TaskType = "replication"
		case "rl-experiment":
			project.TaskType = "reinforcement_learning"
			project.Metric = model.MetricDef{
				Name:      "mean_reward",
				Direction: "higher_is_better",
			}
		}

		if err := s.Init(project); err != nil {
			return fmt.Errorf("initializing marrow: %w", err)
		}

		fmt.Println("Initialized .marrow/ project")
		if initTemplate != "" {
			fmt.Printf("  template: %s\n", initTemplate)
		}
		fmt.Println("  Edit .marrow/marrow.yaml to configure your project.")
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initTemplate, "template", "", "Project template (kaggle-tabular, llm-finetune, paper-replication, rl-experiment)")
}
