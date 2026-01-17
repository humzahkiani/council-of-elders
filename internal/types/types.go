package types

import "time"

// Solution represents an agent's proposed solution to the task
type Solution struct {
	AgentID   int       `json:"agent_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Critique represents an agent's critique of all solutions
type Critique struct {
	AgentID   int       `json:"agent_id"`
	Round     int       `json:"round"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Vote represents an agent's ranked-choice vote
type Vote struct {
	VoterID   int    `json:"voter_id"`
	Rankings  []int  `json:"rankings"`  // Ordered list of AgentIDs, best first (excludes self)
	Reasoning string `json:"reasoning"` // Agent's explanation for their vote
}

// Session represents a complete council session
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

// Config holds CLI configuration
type Config struct {
	AgentCount int
	Rounds     int
	Save       bool
	OutputPath string
	Verbose    bool
	Model      string
	Task       string
}
