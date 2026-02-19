# Marrow

Structured experiment tracking for AI research. A `.marrow/` directory inside your project, plain YAML files, a CLI for you and an MCP server for your AI agent. No database, no cloud, nothing to sign up for.

Sort of like **git for experiment context** — except instead of code it tracks what you tried, what worked, what failed and why.

## Why this exists

Anyone who's done Kaggle competitions or ML research for more than a couple weeks knows the feeling. You tried something last Tuesday that completely tanked your score but now you can't remember what it was or why. Your agent suggests the exact same failed approach for the third time. Your `CLAUDE.md` has devolved into 400 lines of append-only stream of consciousness that nobody — human or AI — can actually parse anymore.

Marrow came out of that frustration. The core idea: experiments should form a DAG not a flat list, learnings should be classified (proven vs assumptions), things that didn't work should go to a graveyard so nobody wastes time on them again, and the index should auto-compute which experiment chain actually led to your best result.

The MCP server is honestly the main point. Your agent can query all of this directly and write back to it — log experiments, record learnings, check what's been tried before. No more copy-pasting context around.

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
mv marrow ~/.local/bin/   # or wherever your PATH points
```

## Quick start

```bash
cd my-kaggle-project
marrow init --template kaggle-tabular

# log some experiments
marrow exp new --model xgboost --metric 0.821 --status neutral --notes "baseline, default params"
marrow exp new --model xgboost --metric 0.842 --status improved --parents exp_001 --tags lr_tuning --notes "lr=0.01, depth=6"
marrow exp new --model xgboost --metric 0.856 --status improved --parents exp_002 --tags feature_eng --notes "added time features"

# record what you learned
marrow learn add "Time-based features improve AUC by 1-2%" --type proven --tags feature_eng
marrow learn add "Ensemble with LightGBM might help" --type assumption --tags ensemble

# record what didn't work (the graveyard)
marrow learn graveyard --approach "LSTM on raw sequences" --reason "OOM even with batch_size=1" --exp exp_002

# see where things stand
marrow summary
```

```
Project: untitled
Task: classification | Metric: AUC-ROC (higher_is_better)

Experiments: 3
Best: exp_003 (AUC-ROC = 0.8560)
Chain: exp_001 → exp_002 → exp_003
Proven: 1 | Assumptions: 1 | Graveyard: 1
```

Everything lives in `.marrow/` — YAML files, one per experiment, human-readable, git-friendly. Diffs and merges just work.

## CLI Reference

### Project setup

```bash
marrow init --template kaggle-tabular    # classification + AUC-ROC
marrow init --template llm-finetune      # generation + eval_loss
marrow init --template paper-replication
marrow init --template rl-experiment     # RL + mean_reward

# or just plain init and edit .marrow/marrow.yaml yourself
marrow init
```

### Experiments

```bash
# log a new experiment
marrow exp new \
  --model xgboost \
  --metric 0.856 \
  --status improved \          # improved | degraded | neutral | failed
  --parents exp_002 \          # which experiment this builds on
  --tags feature_eng,ensemble \
  --notes "added stacking layer"

# list experiments
marrow exp list
marrow exp list --status improved --tag feature_eng --limit 5

# full details
marrow exp show exp_003

# edit after the fact
marrow exp edit exp_003 --notes "actually this was stacking, not blending" --tags feature_eng,stacking

# delete (blocked if other experiments reference it as a parent)
marrow exp delete exp_003
```

Experiments support DAG lineage — `--parents` takes comma-separated IDs. Branch from one experiment into two approaches, both point back. The index figures out which branch won.

### Learnings

```bash
# proven finding
marrow learn add "Batch norm before dropout works better" --type proven --tags architecture

# assumption (not fully verified yet)
marrow learn add "Lower LR might fix overfitting on minority class" --type assumption --tags training

marrow learn list
marrow learn delete learn_001

# graveyard — things that failed
marrow learn graveyard \
  --approach "Polynomial feature expansion" \
  --reason "Tripled training time, +0.001 AUC" \
  --exp exp_005 \
  --tags feature_eng

marrow learn graveyard-list
marrow learn graveyard-delete grave_001
```

There's **conflict detection** when adding learnings. If your new proven finding overlaps with something in the graveyard or clashes with an existing assumption, it warns you. It's heuristic-based (tag overlap and word matching, no embeddings) — not perfect but it catches the obvious "wait didn't we already try this" moments.

### Index & Summary

```bash
marrow index rebuild    # full recompute from all experiments + learnings
marrow index show
marrow summary          # project overview + index
```

The index has two sections. **Computed** is auto-derived: best experiment, winning chain through the DAG, tag counts, status breakdown. **Pinned** is stuff you curate manually — `do_not_try`, `deferred`, `data_warnings`, `critical_features`, `notes`. Pinned fields survive rebuilds, they're never overwritten.

### Context files

Drop any `.yaml` file into `.marrow/context/` for freeform context — EDA observations, feature descriptions, data pipeline notes, whatever you want agents to be able to read.

```bash
marrow ctx list
marrow ctx show eda
```

### Snapshots

```bash
marrow snapshot create --name "before-major-refactor"
marrow snapshot list
```

Copies the full `.marrow/` directory (minus snapshots/) as a timestamped backup.

## MCP Server

This is really the point of the whole thing. Run `marrow mcp` to start an MCP server over stdio. Agents connect and get 16 structured tools to read and write the knowledge base.

### Setup

#### GitHub Copilot (VS Code)

Add to `.vscode/mcp.json`:

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

Or in `.mcp.json`:

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

Settings → MCP Servers → Add, or in `.cursor/mcp.json`:

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

### Tools

Organized roughly by cost, so agents can grab just what they need:

#### Read tools

| Tool | What it does | Typical tokens |
|------|-------------|---------------|
| `get_project_summary` | Project config + index overview. **Start here.** | ~500 |
| `get_best_experiment` | Current best experiment | ~50–200 |
| `get_experiment` | Specific experiment by ID | ~100–300 |
| `get_learnings` | Proven and/or assumptions, filterable by type | ~100–500 |
| `get_failures` | Graveyard — everything that didn't work | ~100–400 |
| `get_data_context` | A named context file (eda, features, etc.) | varies |
| `get_changelog` | Recent mutations, filterable by date | ~100–500 |
| `get_experiment_chain` | Best path through the experiment DAG | ~100–400 |
| `get_experiments_by_tag` | Filter experiments by tags | varies |
| `compare_experiments` | Side-by-side two experiments with delta | ~200 |
| `get_all_experiments` | Everything (use `depth=summary`!) | varies |
| `get_prelude` | **Smart retrieval** — give it your intent, it composes the right context | ~300–800 |

#### Write tools

| Tool | What it does |
|------|-------------|
| `log_experiment` | Log a new experiment (auto-updates index + changelog) |
| `add_learning` | Add a proven finding or assumption (runs conflict detection) |
| `add_graveyard_entry` | Record a failed approach |
| `update_pinned` | Edit the pinned index (do_not_try, deferred, data_warnings, etc.) |

#### Depth parameter

Most read tools accept `depth`:

- **`summary`** — one-liner per item. Cheapest.
- **`standard`** — key fields, no reasoning/environment.
- **`full`** — everything. Use sparingly.

Every response includes a `[tokens≈N depth=X]` header so agents can track their context budget.

### Recommended agent workflow

**1. Start of session — orient:**
```
→ get_project_summary
  "3 experiments, best exp_003 at 0.856 AUC-ROC, chain: exp_001 → exp_002 → exp_003"
```

**2. Before doing work — load context:**
```
→ get_prelude(intent="try feature engineering")
  Returns: project summary + EDA context + data warnings + proven learnings
```

Or go specific:
```
→ get_failures()
→ get_learnings(type="proven")
→ get_best_experiment(depth="full")
```

**3. After an experiment — record it:**
```
→ log_experiment(base_model="xgboost", metric_value=0.862, status="improved", parents="exp_003", notes="added stacking layer")
→ add_learning(text="Stacking with LightGBM improves by 0.6%", type="proven", tags="ensemble")
```

**4. Something failed — graveyard it:**
```
→ add_graveyard_entry(approach="Polynomial features", reason="Tripled training time for +0.001", experiment_id="exp_004")
```

**5. Curate guardrails:**
```
→ update_pinned(field="do_not_try", action="add", value="Random forests — plateau at 0.84")
→ update_pinned(field="data_warnings", action="add", value="Column X has 30% nulls after 2024")
```

### The `get_prelude` tool

Probably the most useful tool for agents. Instead of making 5 calls to piece together context, call `get_prelude` with a natural language intent and it figures out what's relevant:

- `"try new feature engineering"` → EDA context, data warnings, proven learnings
- `"tune hyperparameters"` → tuning-tagged experiments, proven learnings
- `"understand failures"` → graveyard, do_not_try list, proven learnings
- `"try a different model architecture"` → best experiment at full depth, proven learnings

Always includes project summary and proven learnings as a baseline, then adds intent-specific stuff on top.

## Design decisions

**File-per-experiment.** Each experiment is its own YAML file. Git diffs show exactly what changed, merges work naturally. Learnings and graveyard entries are kept in single files since they're smaller and change less often.

**DAG not a list.** Experiments reference parents explicitly. Branch from exp_003 into two different approaches? Both point back. The index walks the DAG backward from the best result and computes the winning chain.

**Computed vs Pinned index.** The computed section can be blown away and rebuilt any time — it's fully derived from your experiments. The pinned section is your guardrails: notes, warnings, things to avoid. It never gets touched by recomputes.

**Conflict detection.** Tag overlap plus word matching against the graveyard and opposite-type learnings. No embeddings, no ML, just string matching. Catches the obvious stuff.

**Atomic writes.** Every YAML write goes through a temp file then `os.Rename()`. If the process crashes mid-write you don't get a half-written file.

**Token-aware MCP responses.** Every response reports approximate token count. Agents can stay within budget without having to guess.

## Status

This is early. It works — the test suite covers the main paths and the MCP server handles real agent sessions — but there's no CI pipeline yet, the tests are a Fish shell script (not Go tests), and it hasn't been tested on Windows. Things might change.

If you run into something broken, open an issue.

## License

MIT
