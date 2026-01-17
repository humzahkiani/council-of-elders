package council

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/humzahkiani/council/internal/agent"
)

// Discuss runs a discussion round where each agent critiques all solutions
func (c *Council) Discuss(ctx context.Context, round int) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(c.agents))

	for _, ag := range c.agents {
		wg.Add(1)
		go func(a *agent.Agent) {
			defer wg.Done()

			critique, err := a.Critique(ctx, c.session.Task, c.session.Solutions, round)
			if err != nil {
				errChan <- fmt.Errorf("agent %d: %w", a.ID, err)
				return
			}

			mu.Lock()
			c.session.Critiques = append(c.session.Critiques, *critique)
			mu.Unlock()

			c.PrintVerboseCritique(critique)
		}(ag)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("discussion errors: %v", errs)
	}

	// Sort critiques by agent ID for consistent ordering
	sort.Slice(c.session.Critiques, func(i, j int) bool {
		if c.session.Critiques[i].Round != c.session.Critiques[j].Round {
			return c.session.Critiques[i].Round < c.session.Critiques[j].Round
		}
		return c.session.Critiques[i].AgentID < c.session.Critiques[j].AgentID
	})

	return nil
}
