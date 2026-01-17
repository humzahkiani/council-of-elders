package council

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/humzahkiani/council/internal/agent"
)

// Generate runs all agents in parallel to create solutions
func (c *Council) Generate(ctx context.Context) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, len(c.agents))

	for _, ag := range c.agents {
		wg.Add(1)
		go func(a *agent.Agent) {
			defer wg.Done()

			solution, err := a.GenerateSolution(ctx, c.session.Task)
			if err != nil {
				errChan <- fmt.Errorf("agent %d: %w", a.ID, err)
				return
			}

			mu.Lock()
			c.session.Solutions = append(c.session.Solutions, *solution)
			mu.Unlock()

			c.PrintVerboseSolution(solution)
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
		return fmt.Errorf("generation errors: %v", errs)
	}

	// Sort solutions by agent ID for consistent ordering
	sort.Slice(c.session.Solutions, func(i, j int) bool {
		return c.session.Solutions[i].AgentID < c.session.Solutions[j].AgentID
	})

	return nil
}
