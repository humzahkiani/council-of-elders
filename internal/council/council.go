package council

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/humzahkiani/council/internal/agent"
	"github.com/humzahkiani/council/internal/storage"
	"github.com/humzahkiani/council/internal/types"
)

// Council orchestrates the multi-agent deliberation process
type Council struct {
	config  *types.Config
	client  *agent.Client
	storage *storage.Storage
	agents  []*agent.Agent
	session *types.Session
}

// New creates a new Council instance
func New(config *types.Config) (*Council, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	client := agent.NewClient(apiKey, config.Model)

	var store *storage.Storage
	var err error
	if config.Save || config.OutputPath != "" {
		store, err = storage.New()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize storage: %w", err)
		}
	}

	// Create agents
	agents := make([]*agent.Agent, config.AgentCount)
	for i := 0; i < config.AgentCount; i++ {
		agents[i] = agent.New(i+1, config.AgentCount, client)
	}

	// Initialize session
	session := &types.Session{
		ID:         uuid.New().String(),
		Task:       config.Task,
		AgentCount: config.AgentCount,
		Rounds:     config.Rounds,
		Model:      config.Model,
		Solutions:  []types.Solution{},
		Critiques:  []types.Critique{},
		Votes:      []types.Vote{},
		Scores:     make(map[int]int),
		CreatedAt:  time.Now(),
	}

	return &Council{
		config:  config,
		client:  client,
		storage: store,
		agents:  agents,
		session: session,
	}, nil
}

// Run executes the full council process: generate -> discuss -> vote -> tally
func (c *Council) Run(ctx context.Context) error {
	c.printHeader()

	// Phase 1: Generate solutions
	c.printPhase("Generating solutions")
	if err := c.Generate(ctx); err != nil {
		return fmt.Errorf("generation phase failed: %w", err)
	}
	c.printPhaseDone()

	// Phase 2: Discussion rounds
	for round := 1; round <= c.config.Rounds; round++ {
		c.printPhase(fmt.Sprintf("Discussion round %d", round))
		if err := c.Discuss(ctx, round); err != nil {
			return fmt.Errorf("discussion phase failed: %w", err)
		}
		c.printPhaseDone()
	}

	// Phase 3: Voting
	c.printPhase("Voting")
	if err := c.Vote(ctx); err != nil {
		return fmt.Errorf("voting phase failed: %w", err)
	}
	c.printPhaseDone()

	// Phase 4: Tally
	c.Tally()
	c.session.CompletedAt = time.Now()

	// Save session if requested
	if err := c.saveSession(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to save session: %v\n", err)
	}

	return nil
}

// Output prints the final results
func (c *Council) Output() {
	fmt.Println()
	fmt.Println("Results")
	fmt.Println("-------")

	// Sort agent IDs for consistent output
	var agentIDs []int
	for id := range c.session.Scores {
		agentIDs = append(agentIDs, id)
	}
	sort.Ints(agentIDs)

	for _, id := range agentIDs {
		score := c.session.Scores[id]
		marker := ""
		if c.session.WinnerID != nil && *c.session.WinnerID == id {
			marker = " * WINNER"
		}
		fmt.Printf("Agent %d: %d points%s\n", id, score, marker)
	}

	fmt.Println()

	if c.session.IsTie {
		fmt.Printf("TIE between Agents %v\n\n", c.session.TiedAgents)
		fmt.Println("All tied solutions are shown below for your review:")
		for _, sol := range c.session.Solutions {
			for _, id := range c.session.TiedAgents {
				if sol.AgentID == id {
					fmt.Printf("\n--- Solution (Agent %d) ---\n%s\n", sol.AgentID, sol.Content)
				}
			}
		}
	} else if c.session.WinnerID != nil {
		fmt.Printf("Winning Solution (Agent %d)\n", *c.session.WinnerID)
		fmt.Println("--------------------------")
		for _, sol := range c.session.Solutions {
			if sol.AgentID == *c.session.WinnerID {
				fmt.Println(sol.Content)
				break
			}
		}
	}
}

// printHeader prints the initial header
func (c *Council) printHeader() {
	fmt.Println("Council of Elders")
	fmt.Println("====================")
	fmt.Printf("Task: %s\n", c.session.Task)
	fmt.Printf("Agents: %d | Rounds: %d | Model: %s\n\n", c.config.AgentCount, c.config.Rounds, c.config.Model)
}

// printPhase prints a phase status
func (c *Council) printPhase(phase string) {
	fmt.Printf("%s... ", phase)
}

// printPhaseDone prints phase completion
func (c *Council) printPhaseDone() {
	fmt.Println("done")
}

// saveSession saves the session if configured
func (c *Council) saveSession() error {
	if c.storage == nil {
		return nil
	}

	var path string
	var err error

	if c.config.OutputPath != "" {
		err = c.storage.SaveTo(c.session, c.config.OutputPath)
		path = c.config.OutputPath
	} else if c.config.Save {
		path, err = c.storage.Save(c.session)
	}

	if err != nil {
		return err
	}

	if path != "" {
		fmt.Printf("\nSession saved to: %s\n", path)
	}

	return nil
}

// PrintVerboseSolution prints a solution in verbose mode
func (c *Council) PrintVerboseSolution(sol *types.Solution) {
	if c.config.Verbose {
		fmt.Printf("\n--- Agent %d Solution ---\n%s\n", sol.AgentID, sol.Content)
	}
}

// PrintVerboseCritique prints a critique in verbose mode
func (c *Council) PrintVerboseCritique(crit *types.Critique) {
	if c.config.Verbose {
		fmt.Printf("\n--- Agent %d Critique (Round %d) ---\n%s\n", crit.AgentID, crit.Round, crit.Content)
	}
}

// PrintVerboseVote prints a vote in verbose mode
func (c *Council) PrintVerboseVote(vote *types.Vote) {
	if c.config.Verbose {
		fmt.Printf("\n--- Agent %d Vote ---\nRankings: %v\nReasoning: %s\n", vote.VoterID, vote.Rankings, vote.Reasoning)
	}
}
