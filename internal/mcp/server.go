package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rzzdr/marrow/internal/store"
)

func NewServer(s *store.Store) *server.MCPServer {
	srv := server.NewMCPServer(
		"marrow",
		"0.1.0",
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
		server.WithInstructions(`Marrow is a structured knowledge base for AI research experiments.
Use get_project_summary for a quick overview. Escalate to deeper tools only when needed.
Prefer summary depth for listings, full depth only for specific experiments.`),
	)

	h := &handlers{store: s}

	srv.AddTool(
		mcp.NewTool("get_project_summary",
			mcp.WithDescription("Get project config + index overview. Call this first in every session. ~500 tokens."),
		),
		h.getProjectSummary,
	)

	srv.AddTool(
		mcp.NewTool("get_best_experiment",
			mcp.WithDescription("Get the current best experiment."),
			mcp.WithString("depth", mcp.Description("summary|standard|full"), mcp.DefaultString("standard")),
		),
		h.getBestExperiment,
	)

	srv.AddTool(
		mcp.NewTool("get_experiment",
			mcp.WithDescription("Get a specific experiment by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Experiment ID (e.g. exp_001)")),
			mcp.WithString("depth", mcp.Description("summary|standard|full"), mcp.DefaultString("full")),
		),
		h.getExperiment,
	)

	srv.AddTool(
		mcp.NewTool("get_learnings",
			mcp.WithDescription("Get proven findings and/or assumptions."),
			mcp.WithString("type", mcp.Description("proven|assumption|all"), mcp.DefaultString("all")),
			mcp.WithString("depth", mcp.Description("summary|standard|full"), mcp.DefaultString("summary")),
		),
		h.getLearnings,
	)

	srv.AddTool(
		mcp.NewTool("get_failures",
			mcp.WithDescription("Get the graveyard — failed approaches and why they failed."),
			mcp.WithString("depth", mcp.Description("summary|standard|full"), mcp.DefaultString("summary")),
		),
		h.getFailures,
	)

	srv.AddTool(
		mcp.NewTool("get_data_context",
			mcp.WithDescription("Get a named context file (e.g. eda, features)."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Context file name (without .yaml)")),
		),
		h.getDataContext,
	)

	srv.AddTool(
		mcp.NewTool("get_changelog",
			mcp.WithDescription("Get recent mutations. Useful to see what changed since last session."),
			mcp.WithString("since", mcp.Description("ISO date to filter from (e.g. 2025-11-13). Empty = all.")),
		),
		h.getChangelog,
	)

	srv.AddTool(
		mcp.NewTool("get_experiment_chain",
			mcp.WithDescription("Get the best experiment chain (DAG walk from root to best)."),
			mcp.WithString("depth", mcp.Description("summary|standard|full"), mcp.DefaultString("summary")),
		),
		h.getExperimentChain,
	)

	srv.AddTool(
		mcp.NewTool("get_experiments_by_tag",
			mcp.WithDescription("Get experiments matching specific tags."),
			mcp.WithString("tags", mcp.Required(), mcp.Description("Comma-separated tags to filter by")),
			mcp.WithString("depth", mcp.Description("summary|standard|full"), mcp.DefaultString("summary")),
		),
		h.getExperimentsByTag,
	)

	srv.AddTool(
		mcp.NewTool("compare_experiments",
			mcp.WithDescription("Compare two experiments side by side."),
			mcp.WithString("id1", mcp.Required(), mcp.Description("First experiment ID")),
			mcp.WithString("id2", mcp.Required(), mcp.Description("Second experiment ID")),
		),
		h.compareExperiments,
	)

	srv.AddTool(
		mcp.NewTool("get_all_experiments",
			mcp.WithDescription("Get all experiments. Can be expensive. Use depth=summary to minimize tokens."),
			mcp.WithString("depth", mcp.Description("summary|standard|full"), mcp.DefaultString("summary")),
		),
		h.getAllExperiments,
	)

	srv.AddTool(
		mcp.NewTool("log_experiment",
			mcp.WithDescription("Log a new experiment result."),
			mcp.WithString("base_model", mcp.Description("Model family (e.g. xgboost, resnet)")),
			mcp.WithString("parents", mcp.Description("Comma-separated parent experiment IDs")),
			mcp.WithNumber("metric_value", mcp.Required(), mcp.Description("Primary metric value")),
			mcp.WithString("status", mcp.Required(), mcp.Description("improved|degraded|neutral|failed")),
			mcp.WithString("tags", mcp.Description("Comma-separated tags")),
			mcp.WithString("notes", mcp.Description("Freeform notes about this experiment")),
		),
		h.logExperiment,
	)

	srv.AddTool(
		mcp.NewTool("add_learning",
			mcp.WithDescription("Add a learning (proven finding or assumption)."),
			mcp.WithString("text", mcp.Required(), mcp.Description("The learning text")),
			mcp.WithString("type", mcp.Required(), mcp.Description("proven|assumption")),
			mcp.WithString("tags", mcp.Description("Comma-separated tags")),
		),
		h.addLearning,
	)

	srv.AddTool(
		mcp.NewTool("add_graveyard_entry",
			mcp.WithDescription("Record a failed approach in the graveyard."),
			mcp.WithString("approach", mcp.Required(), mcp.Description("The approach that failed")),
			mcp.WithString("reason", mcp.Required(), mcp.Description("Why it failed")),
			mcp.WithString("experiment_id", mcp.Description("Related experiment ID")),
			mcp.WithString("tags", mcp.Description("Comma-separated tags")),
		),
		h.addGraveyardEntry,
	)

	srv.AddTool(
		mcp.NewTool("update_pinned",
			mcp.WithDescription("Update pinned index entries (do_not_try, deferred, data_warnings, critical_features, notes)."),
			mcp.WithString("field", mcp.Required(), mcp.Description("do_not_try|deferred|data_warnings|critical_features|notes")),
			mcp.WithString("action", mcp.Required(), mcp.Description("add|remove|set (set replaces all)")),
			mcp.WithString("value", mcp.Required(), mcp.Description("Value to add/remove/set")),
		),
		h.updatePinned,
	)

	srv.AddTool(
		mcp.NewTool("get_prelude",
			mcp.WithDescription("Get an optimized context blob for a given intent. Returns project summary + relevant context based on what you're trying to do."),
			mcp.WithString("intent", mcp.Required(), mcp.Description("What you're about to do (e.g. 'try new feature engineering', 'tune hyperparameters', 'understand failures')")),
		),
		h.getPrelude,
	)

	return srv
}

func Serve(s *store.Store) error {
	srv := NewServer(s)
	return server.ServeStdio(srv)
}
