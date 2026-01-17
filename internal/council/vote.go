package council

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/humzahkiani/council/internal/agent"
	"github.com/humzahkiani/council/internal/types"
)

// Vote collects ranked votes from all agents
func (c *Council) Vote(ctx context.Context) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(c.agents))

	for _, ag := range c.agents {
		wg.Add(1)
		go func(a *agent.Agent) {
			defer wg.Done()

			vote, err := a.Vote(ctx, c.session.Task, c.session.Solutions, c.session.Critiques)
			if err != nil {
				// Per spec: re-prompt agent once, then use empty vote
				vote, err = a.Vote(ctx, c.session.Task, c.session.Solutions, c.session.Critiques)
				if err != nil {
					// Use empty vote if retry fails
					vote = &types.Vote{
						VoterID:   a.ID,
						Rankings:  []int{},
						Reasoning: "Vote failed to parse",
					}
					errChan <- fmt.Errorf("agent %d: vote failed after retry: %w", a.ID, err)
				}
			}

			mu.Lock()
			c.session.Votes = append(c.session.Votes, *vote)
			mu.Unlock()

			c.PrintVerboseVote(vote)
		}(ag)
	}

	wg.Wait()
	close(errChan)

	// Log errors but don't fail - we proceed with whatever votes we got
	for err := range errChan {
		if c.config.Verbose {
			fmt.Printf("Warning: %v\n", err)
		}
	}

	// Sort votes by voter ID for consistent ordering
	sort.Slice(c.session.Votes, func(i, j int) bool {
		return c.session.Votes[i].VoterID < c.session.Votes[j].VoterID
	})

	return nil
}

// Tally calculates scores and determines the winner
func (c *Council) Tally() {
	n := len(c.agents)

	// Initialize scores for all agents
	for _, ag := range c.agents {
		c.session.Scores[ag.ID] = 0
	}

	// Calculate points: 1st place = (N-1) points, 2nd = (N-2), etc.
	for _, vote := range c.session.Votes {
		for i, agentID := range vote.Rankings {
			points := n - 1 - i
			if points > 0 {
				c.session.Scores[agentID] += points
			}
		}
	}

	// Find winner(s)
	maxScore := 0
	for _, score := range c.session.Scores {
		if score > maxScore {
			maxScore = score
		}
	}

	var winners []int
	for agentID, score := range c.session.Scores {
		if score == maxScore {
			winners = append(winners, agentID)
		}
	}

	sort.Ints(winners)

	if len(winners) == 1 {
		c.session.WinnerID = &winners[0]
		c.session.IsTie = false
		c.session.TiedAgents = []int{}
	} else {
		c.session.WinnerID = nil
		c.session.IsTie = true
		c.session.TiedAgents = winners
	}
}
