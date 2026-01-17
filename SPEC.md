# Council of Elders — Specification

## Overview

Council is a CLI tool that presents a task to multiple Claude agents, facilitates discussion among them, and uses ranked-choice voting to determine the best solution. It includes an interactive TUI for reviewing session history.

## Core Concepts

### Agents
- Independent Claude instances that each generate a solution to the given task
- All agents receive identical system prompts and task descriptions
- Agents are identified by numeric IDs (1, 2, 3, etc.)

### Voting Rules
- Ranked-choice voting: each agent ranks all solutions except their own
- Points awarded: 1st place = (N-1) points, 2nd place = (N-2) points, etc.
- Agent cannot vote for their own solution
- Ties are surfaced to the user (no automatic tie-breaking)

### Phases
1. **Generate** — All agents solve the task in parallel
2. **Discuss** — All agents critique all solutions in parallel
3. **Vote** — All agents submit ranked votes in parallel
4. **Tally** — Points are summed, winner is determined

---

## Architecture

```
council/
├── cmd/
│   └── council/
│       └── main.go              # CLI entry point (run & view subcommands)
│
├── internal/
│   ├── agent/
│   │   ├── agent.go             # Agent struct, system prompts
│   │   └── client.go            # Anthropic API client
│   │
│   ├── council/
│   │   ├── council.go           # Orchestrator
│   │   ├── generate.go          # Phase 1: solution generation
│   │   ├── discuss.go           # Phase 2: critiques
│   │   └── vote.go              # Phase 3: voting & tally
│   │
│   ├── storage/
│   │   └── storage.go           # File-based JSON persistence
│   │
│   ├── tui/
│   │   ├── model.go             # Session viewer TUI model
│   │   ├── list.go              # Session list TUI model
│   │   └── styles.go            # Lipgloss styles
│   │
│   └── types/
│       └── types.go             # Shared data structures
│
├── go.mod
└── go.sum
```

---

## CLI Interface

Council uses subcommands for different operations.

### Run Command

Run a council deliberation on a task.

```bash
# Run with default settings (3 agents, 1 discussion round)
council run "Write a function to check if a number is prime"

# Specify agent count
council run --agents 5 "Design a REST API for a blog"

# Multiple discussion rounds
council run --agents 4 --rounds 2 "Implement a rate limiter"

# Save session to default location (~/.council/sessions/)
council run --save "Your task here"

# Save session to specific file
council run --output ./result.json "Your task here"

# Verbose output (show solutions and discussion as they happen)
council run --verbose "Your task here"
```

#### Run Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--agents` | `-a` | 3 | Number of agents (minimum 3) |
| `--rounds` | `-r` | 1 | Number of discussion rounds |
| `--save` | `-s` | false | Save session to ~/.council/sessions/ |
| `--output` | `-o` | "" | Save session to specific file path |
| `--verbose` | `-v` | false | Print detailed output during execution |
| `--model` | `-m` | "claude-sonnet-4-20250514" | Claude model to use |
| `--help` | `-h` | | Show help |

### View Command

View council sessions in an interactive TUI.

```bash
# List all saved sessions (interactive selection)
council view

# View a specific session by ID (partial match supported)
council view abc123

# View a specific session by file path
council view ./my-session.json
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `ANTHROPIC_API_KEY` | Yes (for `run`) | API key for Claude |

---

## TUI Viewer

The TUI viewer provides an interactive interface for exploring council sessions.

### Tabs

| Tab | Content |
|-----|---------|
| **Solutions** | All agent solutions with scores, winner highlighted |
| **Discussion** | Critiques organized by round and agent |
| **Votes** | Each agent's rankings with point breakdown and reasoning |
| **Results** | Final scores, winner announcement, session metadata |

### Keyboard Controls

| Key | Action |
|-----|--------|
| `1`, `2`, `3`, `4` | Jump to specific tab |
| `Tab`, `→`, `l` | Next tab |
| `Shift+Tab`, `←`, `h` | Previous tab |
| `↓`, `j` | Scroll down |
| `↑`, `k` | Scroll up |
| `q`, `Ctrl+C` | Quit |

### Session List

When running `council view` without arguments, an interactive list shows all saved sessions with:
- Winner or tie status
- Task description (truncated)
- Creation date
- Agent count
- Model used

Use arrow keys to navigate and Enter to select a session.

---

## Data Flow

```
                    Task (string)
                         │
                         ▼
    ┌────────────────────────────────────────┐
    │           Phase 1: Generate            │
    │                                        │
    │   ┌─────────┐ ┌─────────┐ ┌─────────┐  │
    │   │ Agent 1 │ │ Agent 2 │ │ Agent 3 │  │
    │   └────┬────┘ └────┬────┘ └────┬────┘  │
    │        │           │           │       │
    │        ▼           ▼           ▼       │
    │   Solution 1  Solution 2  Solution 3   │
    └────────────────────────────────────────┘
                         │
                         ▼
    ┌────────────────────────────────────────┐
    │           Phase 2: Discuss             │
    │                                        │
    │   Each agent receives all solutions    │
    │   Each agent critiques the others      │
    │                                        │
    │   ┌─────────┐ ┌─────────┐ ┌─────────┐  │
    │   │ Agent 1 │ │ Agent 2 │ │ Agent 3 │  │
    │   └────┬────┘ └────┬────┘ └────┬────┘  │
    │        │           │           │       │
    │        ▼           ▼           ▼       │
    │   Critique 1  Critique 2  Critique 3   │
    └────────────────────────────────────────┘
                         │
              (repeat for N rounds)
                         │
                         ▼
    ┌────────────────────────────────────────┐
    │            Phase 3: Vote               │
    │                                        │
    │   Each agent ranks all OTHER solutions │
    │   (cannot vote for own solution)       │
    │                                        │
    │   Agent 1: [2, 3] → 2 gets 2pts, 3 gets 1pt  │
    │   Agent 2: [3, 1] → 3 gets 2pts, 1 gets 1pt  │
    │   Agent 3: [1, 2] → 1 gets 2pts, 2 gets 1pt  │
    └────────────────────────────────────────┘
                         │
                         ▼
    ┌────────────────────────────────────────┐
    │              Tally                     │
    │                                        │
    │   Agent 1: 3 points                    │
    │   Agent 2: 3 points                    │
    │   Agent 3: 3 points                    │
    │                                        │
    │   Result: Tie (surfaced to user)       │
    └────────────────────────────────────────┘
```

---

## Data Types

### Solution

```go
type Solution struct {
    AgentID   int       `json:"agent_id"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Critique

```go
type Critique struct {
    AgentID   int       `json:"agent_id"`
    Round     int       `json:"round"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Vote

```go
type Vote struct {
    VoterID   int    `json:"voter_id"`
    Rankings  []int  `json:"rankings"`  // Ordered list of AgentIDs, best first
    Reasoning string `json:"reasoning"` // Agent's explanation for their vote
}
```

### Session

```go
type Session struct {
    ID          string         `json:"id"`
    Task        string         `json:"task"`
    AgentCount  int            `json:"agent_count"`
    Rounds      int            `json:"rounds"`
    Model       string         `json:"model"`
    Solutions   []Solution     `json:"solutions"`
    Critiques   []Critique     `json:"critiques"`
    Votes       []Vote         `json:"votes"`
    Scores      map[int]int    `json:"scores"`
    WinnerID    *int           `json:"winner_id"`
    IsTie       bool           `json:"is_tie"`
    TiedAgents  []int          `json:"tied_agents"`
    CreatedAt   time.Time      `json:"created_at"`
    CompletedAt time.Time      `json:"completed_at"`
}
```

### Config

```go
type Config struct {
    AgentCount int
    Rounds     int
    Save       bool
    OutputPath string
    Verbose    bool
    Model      string
    Task       string
}
```

---

## Storage

### Directory Structure

```
~/.council/
└── sessions/
    ├── 2024-01-15_143022_a1b2c3.json
    ├── 2024-01-15_151847_d4e5f6.json
    └── ...
```

### Session Filename Format

```
{date}_{time}_{short_id}.json
YYYY-MM-DD_HHMMSS_{first 6 chars of UUID}.json
```

### Example Session File

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "task": "Write a function to check if a number is prime",
  "agent_count": 3,
  "rounds": 1,
  "model": "claude-sonnet-4-20250514",
  "solutions": [
    {
      "agent_id": 1,
      "content": "Here's my solution using trial division...",
      "created_at": "2024-01-15T14:30:22Z"
    },
    {
      "agent_id": 2,
      "content": "I'll use the 6k±1 optimization...",
      "created_at": "2024-01-15T14:30:24Z"
    },
    {
      "agent_id": 3,
      "content": "Here's an implementation with Miller-Rabin...",
      "created_at": "2024-01-15T14:30:23Z"
    }
  ],
  "critiques": [
    {
      "agent_id": 1,
      "round": 1,
      "content": "Solution 2's optimization is good but Solution 3's Miller-Rabin is overkill...",
      "created_at": "2024-01-15T14:30:45Z"
    }
  ],
  "votes": [
    {
      "voter_id": 1,
      "rankings": [2, 3],
      "reasoning": "Solution 2 provides the best balance of efficiency and readability."
    },
    {
      "voter_id": 2,
      "rankings": [3, 1],
      "reasoning": "Solution 3 handles edge cases well, Solution 1 is a solid fallback."
    },
    {
      "voter_id": 3,
      "rankings": [2, 1],
      "reasoning": "Solution 2's optimization is practical and meaningful."
    }
  ],
  "scores": {
    "1": 2,
    "2": 4,
    "3": 3
  },
  "winner_id": 2,
  "is_tie": false,
  "tied_agents": [],
  "created_at": "2024-01-15T14:30:22Z",
  "completed_at": "2024-01-15T14:31:15Z"
}
```

---

## Agent Prompts

### Solution Generation Prompt

```
You are Agent {N} in a council of {TOTAL} agents. You have been given a task to solve.

Provide your solution to the task. Be thorough but concise. Focus on correctness and clarity.

Do not reference other agents or solutions — you are working independently.
```

### Discussion Prompt

```
You are Agent {N} in a council of {TOTAL} agents.

Review all solutions and provide your critique. For each solution OTHER than your own:
- Identify strengths
- Identify weaknesses or potential issues
- Suggest improvements if applicable

Be constructive and objective. Your goal is to help identify the best solution.
```

### Voting Prompt

```
You are Agent {N} in a council of {TOTAL} agents. You have seen all solutions and the discussion.

Rank all solutions EXCEPT YOUR OWN from best to worst. You CANNOT vote for your own solution (Solution {N}).

Respond with a JSON object in this exact format:
{
  "rankings": [X, Y, ...],
  "reasoning": "Brief explanation of your ranking"
}

Where X is the agent number of your top choice, Y is your second choice, etc.
Do not include your own agent number ({N}) in the rankings.
```

---

## Output Format

### Standard Output (non-verbose)

```
Council of Elders
====================
Task: Write a function to check if a number is prime
Agents: 3 | Rounds: 1 | Model: claude-sonnet-4-20250514

Generating solutions... done
Discussion round 1... done
Voting... done

Results
-------
Agent 1: 2 points
Agent 2: 4 points * WINNER
Agent 3: 3 points

Winning Solution (Agent 2)
--------------------------
I'll use the 6k±1 optimization for checking primality...

[solution content]
```

### Verbose Output

Includes full solutions, critiques, and vote reasoning as they are generated.

### Tie Output

```
Results
-------
Agent 1: 3 points
Agent 2: 3 points
Agent 3: 3 points

TIE between Agents 1, 2, 3

All solutions are shown below for your review:
[solutions]
```

---

## Error Handling

| Error | Behavior |
|-------|----------|
| Missing API key | Exit with message: "ANTHROPIC_API_KEY environment variable not set" |
| API rate limit | Retry with exponential backoff (max 3 retries) |
| API error | Exit with error message, save partial session if --save |
| Invalid agent count (<3) | Exit with message: "Minimum 3 agents required" |
| Invalid vote format | Re-prompt agent once, then use empty vote |

---

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/google/uuid` | UUID generation |
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/lipgloss` | TUI styling |
| `github.com/charmbracelet/bubbles` | TUI components (list, viewport) |

---

## Future Considerations

- **Moderator mode** — Optional moderator agent for 6+ agents
- **Custom prompts** — Allow user-defined system prompts
- **Model mixing** — Different models for different agents
- **Session resume** — Continue a previous session with additional rounds
- **Export formats** — Markdown, HTML output options
- **Claude Code plugin** — MCP server or slash command integration
