# Council of Elders

A CLI tool that runs multiple Claude agents in parallel, facilitates discussion among them, and uses ranked-choice voting to determine the best solution. Built in Go with an interactive TUI for session review.

## Features

- **Parallel Processing** — Multiple agents generate solutions simultaneously
- **Peer Discussion** — Agents critique each other's solutions
- **Ranked-Choice Voting** — Democratic selection with no self-voting
- **Interactive TUI** — Browse and review session history
- **Session Persistence** — Save and reload deliberations as JSON

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/council-of-elders.git
cd council-of-elders

# Build
go build -o council ./cmd/council

# Set your API key
export ANTHROPIC_API_KEY="your-key"
```

## Usage

### Run a Council

```bash
# Basic usage (3 agents, 1 discussion round)
./council run "Write a function to check if a number is prime"

# More agents and discussion rounds
./council run --agents 5 --rounds 2 "Design a REST API for a blog"

# Save session for later review
./council run --save "Your task here"

# Verbose output (show solutions as they're generated)
./council run --verbose "Your task here"
```

#### Run Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--agents` | `-a` | 3 | Number of agents (minimum 3) |
| `--rounds` | `-r` | 1 | Number of discussion rounds |
| `--save` | `-s` | false | Save session to ~/.council/sessions/ |
| `--output` | `-o` | "" | Save session to specific file path |
| `--verbose` | `-v` | false | Print detailed output during execution |
| `--model` | `-m` | claude-sonnet-4-20250514 | Claude model to use |

### View Sessions

```bash
# Interactive session browser
./council view

# View specific session by ID
./council view abc123

# View session from file
./council view ./path/to/session.json
```

#### TUI Controls

| Key | Action |
|-----|--------|
| `1`, `2`, `3`, `4` | Jump to tab (Solutions, Discussion, Votes, Results) |
| `Tab`, `→`, `l` | Next tab |
| `Shift+Tab`, `←`, `h` | Previous tab |
| `↓`, `j` / `↑`, `k` | Scroll down/up |
| `q`, `Ctrl+C` | Quit |

## How It Works

```
User Task → Generate (parallel) → Solutions
         → Discuss (parallel)  → Critiques
         → Vote (parallel)     → Rankings
         → Tally               → Winner/Tie
```

### Phases

1. **Generate** — All agents solve the task independently in parallel
2. **Discuss** — Each agent critiques all other solutions
3. **Vote** — Agents rank solutions (excluding their own) with ranked-choice voting
4. **Tally** — Points summed (1st = N-1 points, 2nd = N-2, etc.), winner determined

### Voting Rules

- Agents cannot vote for their own solution
- Rankings use Borda count: 1st place = (N-1) points, 2nd = (N-2), etc.
- Ties are surfaced to the user (no automatic tie-breaking)

## Project Structure

```
council-of-elders/
├── cmd/council/main.go          # CLI entry point
├── internal/
│   ├── agent/                   # Agent prompts & API client
│   ├── council/                 # Orchestrator (generate, discuss, vote)
│   ├── storage/                 # JSON persistence
│   ├── tui/                     # Interactive viewer
│   └── types/                   # Shared data structures
├── SPEC.md                      # Full specification
├── ARCHITECTURE.md              # Design decisions
└── DEVLOG.md                    # Development log
```

## Requirements

- Go 1.21+
- Anthropic API key

## License

MIT
