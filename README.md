# Marrow

A structured knowledge base for AI research experiments. Think of it as **git for your experiment context** — but instead of tracking code, it tracks what you tried, what worked, what failed, and why.

Marrow sits inside your project as a `.marrow/` directory and gives you two interfaces: a CLI for humans and an MCP server for AI agents. Everything is plain YAML on disk. No database, no cloud, no lock-in.

## Why

If you've done Kaggle competitions, ML research, or any kind of iterative experimentation, you know the pain:

- You tried something last week that didn't work, but you can't remember why
- Your AI agent keeps suggesting approaches you already ruled out
- Context files like `CLAUDE.md` turn into messy append-only dumps
- After 30 experiments, you've lost track of which ones actually mattered

Marrow fixes this by giving structure to experiment history. Experiments form a DAG (not just a flat list), learnings are classified as proven or assumptions, failed approaches go to a graveyard so nobody wastes time on them again, and the index auto-computes which experiment chain actually led to your best result.

The MCP server means your AI agent can query all of this directly — and write back to it. The agent can log experiments, record learnings, and check what's been tried before, all without you copy-pasting context around.

## Install

Requires Go 1.25+.

```bash
go install github.com/rzzdr/marrow/cmd/marrow@latest
```

Or build from source:

```bash
git clone https://github.com/rzzdr/marrow.git
cd marrow
go build -o marrow ./cmd/marrow
# move to somewhere in your PATH
mv marrow ~/.local/bin/
```

## Quick Start

```bash
# Initialize in your project directory
cd my-kaggle-project
marrow init --template kaggle-tabular

# Log some experiments
marrow exp new --model xgboost --metric 0.821 --status neutral --notes "baseline, default params"
marrow exp new --model xgboost --metric 0.842 --status improved --parents exp_001 --tags lr_tuning --notes "lr=0.01, depth=6"
marrow exp new --model xgboost --metric 0.856 --status improved --parents exp_002 --tags feature_eng --notes "added time features"

# Record what you've learned
marrow learn add "Time-based features improve AUC by 1-2%" --type proven --tags feature_eng
marrow learn add "Ensemble with LightGBM might help" --type assumption --tags ensemble

# Record what didn't work
marrow learn graveyard --approach "LSTM on raw sequences" --reason "OOM even with batch_size=1" --exp exp_002

# See where you stand
marrow summary
```

Output:

```
Project: untitled
Task: classification | Metric: AUC-ROC (higher_is_better)

Experiments: 3
Best: exp_003 (AUC-ROC = 0.8560)
Chain: exp_001 → exp_002 → exp_003
Proven: 1 | Assumptions: 1 | Graveyard: 1
```

## What Goes Where

When you run `marrow init`, it creates this structure:

```
your-project/
└── .marrow/
    ├── marrow.yaml            # project config — name, task type, metric
    ├── index.yaml             # auto-computed stats + your pinned notes
    ├── changelog.yaml         # every mutation, timestamped
    ├── experiments/
    │   ├── exp_001.yaml       # one file per experiment
    │   ├── exp_002.yaml
    │   └── ...
    ├── learnings/
    │   ├── learnings.yaml     # proven findings + assumptions
    │   └── graveyard.yaml     # approaches that failed
    ├── context/               # freeform files — eda notes, feature docs, etc.
    └── snapshots/             # point-in-time backups
```

All YAML, all human-readable, all git-friendly. Each experiment is its own file so diffs and merges work naturally.

## CLI Reference

### Project Setup

```bash
# Initialize with a template
marrow init --template kaggle-tabular    # classification + AUC-ROC
marrow init --template llm-finetune      # generation + eval_loss
marrow init --template paper-replication
marrow init --template rl-experiment     # RL + mean_reward

# Or just plain init and edit .marrow/marrow.yaml yourself
marrow init
```

### Experiments

```bash
# Log a new experiment
marrow exp new \
  --model xgboost \
  --metric 0.856 \
  --status improved \          # improved | degraded | neutral | failed
  --parents exp_002 \          # which experiment this builds on (DAG lineage)
  --tags feature_eng,ensemble \
  --notes "added stacking layer"

# List all experiments (one-liners)
marrow exp list

# Full details on a specific experiment
marrow exp show exp_003
```

Experiments support DAG lineage — `--parents` can take multiple comma-separated IDs if your experiment builds on more than one prior.

### Learnings

```bash
# Add a proven finding
marrow learn add "Batch norm before dropout works better" --type proven --tags architecture

# Add an assumption you haven't fully verified
marrow learn add "Lower LR might fix overfitting on minority class" --type assumption --tags training

# List all learnings
marrow learn list

# Record a failed approach
marrow learn graveyard \
  --approach "Polynomial feature expansion" \
  --reason "Tripled training time, +0.001 AUC" \
  --exp exp_005 \
  --tags feature_eng

# List graveyard
marrow learn graveyard-list
```

Marrow runs **conflict detection** when you add a learning — if your new "proven" finding contradicts something in the graveyard, or clashes with an existing assumption, it warns you.

### Index & Summary

```bash
# Rebuild the index from scratch (recomputes best chain, counts, etc.)
marrow index rebuild

# Show the computed index
marrow index show

# Quick overview of project + index
marrow summary
```

The index has two sections:
- **Computed** — auto-derived from your data (best experiment, experiment chain through the DAG, tag counts, etc.)
- **Pinned** — things you curate manually: `do_not_try`, `deferred`, `data_warnings`, `critical_features`, `notes`. These survive rebuilds.

### Context Files

Drop any `.yaml` file into `.marrow/context/` to store freeform context — EDA observations, feature descriptions, data pipeline notes, whatever.

```bash
marrow ctx list          # list available context files
marrow ctx show eda      # print contents of .marrow/context/eda.yaml
```

### Snapshots

```bash
marrow snapshot create --name "before-major-refactor"
marrow snapshot list
```

Copies the entire `.marrow/` directory (minus snapshots/) as a timestamped backup.

## MCP Server — AI Agent Integration

This is the main thing. Run `marrow mcp` to start an MCP server over stdio. AI agents connect to it and get 16 structured tools to read and write your knowledge base.

### Setup by Agent

#### GitHub Copilot (VS Code)

Add to your `.vscode/mcp.json`:

```json
{
  "servers": {
    "marrow": {
      "type": "stdio",
      "command": "marrow",
      "args": ["mcp"]
    }
  }
}
```

#### Claude Code

```bash
claude mcp add marrow -- marrow mcp
```

Or add to your `.mcp.json`:

```json
{
  "mcpServers": {
    "marrow": {
      "command": "marrow",
      "args": ["mcp"]
    }
  }
}
```

#### Cursor

Go to **Settings → MCP Servers → Add**, or add to `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "marrow": {
      "command": "marrow",
      "args": ["mcp"]
    }
  }
}
```

#### Windsurf

Add to `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "marrow": {
      "command": "marrow",
      "args": ["mcp"]
    }
  }
}
```

#### OpenAI Codex CLI

```bash
codex mcp add marrow -- marrow mcp
```

### Available Tools

Tools are organized by cost so agents can fetch just what they need:

#### Read Tools

| Tool | What it does | Typical tokens |
|------|-------------|---------------|
| `get_project_summary` | Project config + index overview. **Call this first.** | ~500 |
| `get_best_experiment` | Current best experiment | ~50–200 |
| `get_experiment` | Specific experiment by ID | ~100–300 |
| `get_learnings` | Proven and/or assumptions (filterable by type) | ~100–500 |
| `get_failures` | Graveyard — everything that didn't work | ~100–400 |
| `get_data_context` | A named context file (eda, features, etc.) | varies |
| `get_changelog` | Recent mutations, filterable by date | ~100–500 |
| `get_experiment_chain` | Best path through the experiment DAG | ~100–400 |
| `get_experiments_by_tag` | Filter experiments by tags | varies |
| `compare_experiments` | Side-by-side two experiments with delta | ~200 |
| `get_all_experiments` | Everything (use `depth=summary`!) | varies |
| `get_prelude` | **Smart retrieval** — give it your intent and it composes the right context | ~300–800 |

#### Write Tools

| Tool | What it does |
|------|-------------|
| `log_experiment` | Log a new experiment (auto-updates index + changelog) |
| `add_learning` | Add a proven finding or assumption (runs conflict detection) |
| `add_graveyard_entry` | Record a failed approach |
| `update_pinned` | Edit the pinned index (do_not_try, deferred, data_warnings, etc.) |

#### Depth Parameter

Most read tools accept a `depth` parameter:

- **`summary`** — one-liner per item. Cheapest. Good for listings.
- **`standard`** — key fields, no reasoning text or environment details.
- **`full`** — everything. Use only when you need to deeply inspect one experiment.

Every response includes a `[tokens≈N depth=X]` header so agents can track context budget.

### Recommended Agent Workflow

Here's how an agent should use Marrow in practice:

**1. Start of session — orient yourself:**
```
→ get_project_summary
  "3 experiments, best exp_003 at 0.856 AUC-ROC, chain: exp_001 → exp_002 → exp_003"
```

**2. Before doing work — load relevant context:**
```
→ get_prelude(intent="try feature engineering")
  Returns: project summary + EDA context + data warnings + proven learnings
```

Or be specific:
```
→ get_failures()           # what NOT to try
→ get_learnings(type="proven")  # what works
→ get_best_experiment(depth="full")  # current best in detail
```

**3. After running an experiment — record results:**
```
→ log_experiment(base_model="xgboost", metric_value=0.862, status="improved", parents="exp_003", notes="added stacking layer")
→ add_learning(text="Stacking with LightGBM improves by 0.6%", type="proven", tags="ensemble")
```

**4. If something failed — save it to the graveyard:**
```
→ add_graveyard_entry(approach="Polynomial features", reason="Tripled training time for +0.001", experiment_id="exp_004")
```

**5. Curate guardrails for future sessions:**
```
→ update_pinned(field="do_not_try", action="add", value="Random forests — plateau at 0.84")
→ update_pinned(field="data_warnings", action="add", value="Column X has 30% nulls after 2024")
```

### The `get_prelude` Tool

This is the most powerful tool for agents. Instead of making 5 calls to piece together context, call `get_prelude` with a natural language intent:

- `"try new feature engineering"` → returns EDA context, data warnings, proven learnings
- `"tune hyperparameters"` → returns experiments tagged with tuning, proven learnings
- `"understand failures"` → returns graveyard, do_not_try list, proven learnings
- `"try a different model architecture"` → returns best experiment in full detail, proven learnings

It always includes the project summary and proven learnings, then adds intent-specific context on top.

## Design Choices

**File-per-experiment** — each experiment is its own YAML file. Git diffs show exactly what changed, merges work naturally, and you never parse a giant monolith.

**DAG lineage** — experiments reference parents, not just a sequential number. Branch out from exp_003 to try two different approaches? Both point back to exp_003. The index figures out which branch won.

**Computed + Pinned index** — the computed section rebuilds from scratch any time you run `index rebuild`. The pinned section is yours — notes, warnings, things to avoid. It never gets overwritten by recomputes.

**Conflict detection** — adding a learning that contradicts a graveyard entry or an opposite-type learning triggers a warning. Catches the "wait, didn't we already try this?" moments.

**Atomic writes** — YAML writes go through a temp file + rename. If the process crashes mid-write, you don't get a corrupted file.

**Token-aware** — every MCP response reports approximate token count. Agents can stay within budget without guessing.

## License

MIT
