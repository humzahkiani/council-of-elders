# Council of Elders — Development Log

> Session-by-session record of development progress. For architecture and design docs, see [ARCHITECTURE.md](./ARCHITECTURE.md).

---

## Session 1 — 2026-01-16

### Goals
- Design and implement the initial Council tool
- Create a TUI for viewing session history

### What Was Done

1. **Designed the system**
   - Discussed 3 approaches for multi-agent deliberation (sequential, moderator, tournament)
   - Chose parallel generation with optional moderator
   - Designed voting system (ranked-choice, no self-voting)
   - Chose Go for implementation
   - Decided on file-based JSON storage

2. **Built core CLI**
   - `council run "task"` — Run a deliberation
   - `council view` — Browse/view sessions in TUI
   - Flags: --agents, --rounds, --save, --output, --verbose, --model

3. **Built agent system**
   - Anthropic API client with retry logic
   - Agent prompts for generate/discuss/vote phases
   - JSON extraction for vote parsing

4. **Built council orchestrator**
   - Parallel solution generation
   - Parallel discussion/critiques
   - Parallel voting + tally
   - Session persistence

5. **Built TUI viewer**
   - Session list with interactive selection
   - Session viewer with 4 tabs (Solutions, Discussion, Votes, Results)
   - Keyboard navigation

6. **Documentation**
   - Created SPEC.md
   - Created sample session for TUI testing

### Files Created

```
cmd/council/main.go
internal/agent/agent.go
internal/agent/client.go
internal/council/council.go
internal/council/generate.go
internal/council/discuss.go
internal/council/vote.go
internal/storage/storage.go
internal/tui/model.go
internal/tui/list.go
internal/tui/styles.go
internal/types/types.go
go.mod
.gitignore
SPEC.md
```

### Design Decisions

- **No moderator by default**: For 3-5 agents, direct peer discussion works well
- **TUI over web UI**: Faster to build, terminal-native, no server overhead
- **Ranked-choice voting**: More expressive, surfaces consensus strength
- **File-based storage**: Simple, human-readable, no dependencies

### Current State

- Build compiles successfully
- CLI runs with all subcommands
- TUI displays and navigates
- User validated core functionality works with real API

### Issues / TODOs

- [ ] No tests yet
- [ ] No progress indicator during API calls
- [ ] Vote parsing edge cases

### Next Steps

- Add tests
- Consider Claude Code plugin (MCP server)
- Add progress indicators

---

## Session Template

```markdown
## Session N — YYYY-MM-DD

### Goals
-

### What Was Done
-

### Files Changed
-

### Design Decisions
-

### Issues Encountered
-

### Next Steps
-
```
