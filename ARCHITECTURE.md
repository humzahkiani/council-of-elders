# Council of Elders — Architecture

## Overview

A CLI tool that runs multiple Claude agents in parallel, facilitates discussion among them, and uses ranked-choice voting to determine the best solution. Built in Go with an interactive TUI for session review.

---

## Project Structure

```
council-of-ai-elders/
├── cmd/council/main.go          # CLI entry point (run & view subcommands)
├── internal/
│   ├── agent/
│   │   ├── agent.go             # Agent struct, prompts, vote parsing
│   │   └── client.go            # Anthropic API client with retry
│   ├── council/
│   │   ├── council.go           # Main orchestrator
│   │   ├── generate.go          # Phase 1: parallel solution generation
│   │   ├── discuss.go           # Phase 2: parallel critiques
│   │   └── vote.go              # Phase 3: voting + tally
│   ├── storage/
│   │   └── storage.go           # JSON file persistence
│   ├── tui/
│   │   ├── model.go             # Session viewer (4 tabs)
│   │   ├── list.go              # Session list browser
│   │   └── styles.go            # Lipgloss styling
│   └── types/
│       └── types.go             # Shared data structures
├── go.mod
├── go.sum
├── .gitignore
├── SPEC.md                      # Full specification
├── ARCHITECTURE.md              # This file
└── DEVLOG.md                    # Session-by-session development log
```

---

## Data Flow

```
User Task → Generate (parallel) → Solutions
         → Discuss (parallel)  → Critiques
         → Vote (parallel)     → Votes
         → Tally               → Winner/Tie
         → Save (optional)     → JSON file
         → View (TUI)          → Interactive review
```

---

## Design Decisions

### Why Go?

- **Goroutines**: Natural fit for parallel API calls
- **Single binary**: Easy distribution, no runtime dependencies
- **Strong typing**: Catches errors at compile time
- **Fast startup**: Important for CLI tools

### Why No Moderator (Default)?

- For 3-5 agents, direct peer discussion works well
- Avoids single point of failure/bias
- Simpler implementation
- Moderator can be added later for 6+ agent scenarios

### Why Ranked-Choice Voting?

- More expressive than single-vote
- Surfaces consensus strength (unanimous vs. split)
- No self-voting prevents bias
- Ties are informative, not failures

### Why TUI Over Web UI?

- Faster to build
- No server/browser overhead
- Terminal-native workflow
- Can add web UI later if needed

### Why File-Based Storage?

- Simple, no database dependencies
- Human-readable JSON
- Easy to backup/share sessions
- SQLite could be added later for querying

---

## Concurrency Pattern

All phases use the same goroutine pattern:

```go
var wg sync.WaitGroup
var mu sync.Mutex
errChan := make(chan error, len(agents))

for _, agent := range agents {
    wg.Add(1)
    go func(a *Agent) {
        defer wg.Done()
        result, err := a.DoWork(ctx)
        if err != nil {
            errChan <- err
            return
        }
        mu.Lock()
        results = append(results, result)
        mu.Unlock()
    }(agent)
}

wg.Wait()
close(errChan)
```

---

## API Client

### Endpoint
```
POST https://api.anthropic.com/v1/messages
```

### Headers
```
x-api-key: {ANTHROPIC_API_KEY}
anthropic-version: 2023-06-01
content-type: application/json
```

### Retry Logic
- Retry on HTTP 429 (rate limit)
- Exponential backoff: 1s, 2s, 4s
- Max 3 retries

---

## Vote Parsing

Agents return JSON like:
```json
{
  "rankings": [2, 3],
  "reasoning": "Agent 2's solution is more efficient..."
}
```

Parser handles:
- JSON wrapped in markdown code blocks (```json ... ```)
- Extracting first `{...}` from response text
- Validation: no self-voting, valid agent IDs

---

## Scoring Algorithm

```go
for _, vote := range votes {
    for i, agentID := range vote.Rankings {
        points := numAgents - 1 - i  // 1st = N-1, 2nd = N-2, etc.
        scores[agentID] += points
    }
}
```

Example with 3 agents:
- 1st place = 2 points
- 2nd place = 1 point
- (Can't vote for self, so only 2 rankings per agent)

---

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/spf13/cobra | v1.8.0 | CLI framework |
| github.com/google/uuid | v1.6.0 | UUID generation |
| github.com/charmbracelet/bubbletea | v1.3.10 | TUI framework |
| github.com/charmbracelet/lipgloss | v1.1.0 | TUI styling |
| github.com/charmbracelet/bubbles | v0.21.0 | TUI components |

---

## Development Commands

```bash
# Build
go build -o council ./cmd/council

# Run a council
export ANTHROPIC_API_KEY="your-key"
./council run --save "Your task here"

# Run with options
./council run --agents 5 --rounds 2 --verbose "Your task"

# View sessions (interactive list)
./council view

# View specific session
./council view abc123
./council view ./path/to/session.json
```

---

## Future Roadmap

### Near Term
- [ ] Add unit tests
- [ ] Add integration tests
- [ ] Progress indicators during API calls
- [ ] Better error messages
- [ ] Loading state in TUI

### Medium Term
- [ ] Moderator mode for 6+ agents
- [ ] Claude Code plugin (MCP server or slash command)
- [ ] Custom system prompts via config

### Long Term
- [ ] Model mixing (different models per agent)
- [ ] Session resume/continuation
- [ ] Export to Markdown/HTML
- [ ] Web UI option
