package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/humzahkiani/council/internal/types"
)

// Agent represents a single Claude instance in the council
type Agent struct {
	ID     int
	Total  int
	client *Client
}

// New creates a new agent
func New(id, total int, client *Client) *Agent {
	return &Agent{
		ID:     id,
		Total:  total,
		client: client,
	}
}

// GenerateSolution creates a solution for the given task
func (a *Agent) GenerateSolution(ctx context.Context, task string) (*types.Solution, error) {
	system := a.generationPrompt()
	messages := []Message{
		{Role: "user", Content: task},
	}

	response, err := a.client.SendMessage(ctx, system, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate solution: %w", err)
	}

	return &types.Solution{
		AgentID:   a.ID,
		Content:   response,
		CreatedAt: time.Now(),
	}, nil
}

// Critique generates critiques of all solutions
func (a *Agent) Critique(ctx context.Context, task string, solutions []types.Solution, round int) (*types.Critique, error) {
	system := a.discussionPrompt()
	userContent := a.formatDiscussionRequest(task, solutions)
	messages := []Message{
		{Role: "user", Content: userContent},
	}

	response, err := a.client.SendMessage(ctx, system, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate critique: %w", err)
	}

	return &types.Critique{
		AgentID:   a.ID,
		Round:     round,
		Content:   response,
		CreatedAt: time.Now(),
	}, nil
}

// Vote ranks all solutions except the agent's own
func (a *Agent) Vote(ctx context.Context, task string, solutions []types.Solution, critiques []types.Critique) (*types.Vote, error) {
	system := a.votingPrompt()
	userContent := a.formatVotingRequest(task, solutions, critiques)
	messages := []Message{
		{Role: "user", Content: userContent},
	}

	response, err := a.client.SendMessage(ctx, system, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate vote: %w", err)
	}

	return a.parseVote(response)
}

// generationPrompt returns the system prompt for solution generation
func (a *Agent) generationPrompt() string {
	return fmt.Sprintf(`You are Agent %d in a council of %d agents. You have been given a task to solve.

Provide your solution to the task. Be thorough but concise. Focus on correctness and clarity.

Do not reference other agents or solutions — you are working independently.`, a.ID, a.Total)
}

// discussionPrompt returns the system prompt for discussion/critique
func (a *Agent) discussionPrompt() string {
	return fmt.Sprintf(`You are Agent %d in a council of %d agents.

Review all solutions and provide your critique. For each solution OTHER than your own:
- Identify strengths
- Identify weaknesses or potential issues
- Suggest improvements if applicable

Be constructive and objective. Your goal is to help identify the best solution.`, a.ID, a.Total)
}

// votingPrompt returns the system prompt for voting
func (a *Agent) votingPrompt() string {
	return fmt.Sprintf(`You are Agent %d in a council of %d agents. You have seen all solutions and the discussion.

Rank all solutions EXCEPT YOUR OWN from best to worst. You CANNOT vote for your own solution (Solution %d).

Respond with a JSON object in this exact format:
{
  "rankings": [X, Y, ...],
  "reasoning": "Brief explanation of your ranking"
}

Where X is the agent number of your top choice, Y is your second choice, etc.
Do not include your own agent number (%d) in the rankings.`, a.ID, a.Total, a.ID, a.ID)
}

// formatDiscussionRequest formats the user message for discussion
func (a *Agent) formatDiscussionRequest(task string, solutions []types.Solution) string {
	var sb strings.Builder
	sb.WriteString("## Task\n")
	sb.WriteString(task)
	sb.WriteString("\n\n## Solutions\n\n")

	for _, sol := range solutions {
		sb.WriteString(fmt.Sprintf("### Solution %d (Agent %d)\n", sol.AgentID, sol.AgentID))
		sb.WriteString(sol.Content)
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// formatVotingRequest formats the user message for voting
func (a *Agent) formatVotingRequest(task string, solutions []types.Solution, critiques []types.Critique) string {
	var sb strings.Builder
	sb.WriteString("## Task\n")
	sb.WriteString(task)
	sb.WriteString("\n\n## Solutions\n\n")

	for _, sol := range solutions {
		sb.WriteString(fmt.Sprintf("### Solution %d (Agent %d)\n", sol.AgentID, sol.AgentID))
		sb.WriteString(sol.Content)
		sb.WriteString("\n\n")
	}

	if len(critiques) > 0 {
		sb.WriteString("## Discussion\n\n")
		for _, crit := range critiques {
			sb.WriteString(fmt.Sprintf("### Agent %d's Critique\n", crit.AgentID))
			sb.WriteString(crit.Content)
			sb.WriteString("\n\n")
		}
	}

	sb.WriteString(fmt.Sprintf("Now provide your vote. Remember: you are Agent %d and cannot vote for your own solution.\n", a.ID))

	return sb.String()
}

// voteResponse represents the expected JSON structure of a vote
type voteResponse struct {
	Rankings  []int  `json:"rankings"`
	Reasoning string `json:"reasoning"`
}

// parseVote extracts and validates a vote from the agent's response
func (a *Agent) parseVote(response string) (*types.Vote, error) {
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var voteResp voteResponse
	if err := json.Unmarshal([]byte(jsonStr), &voteResp); err != nil {
		return nil, fmt.Errorf("failed to parse vote JSON: %w", err)
	}

	// Validate: ensure agent did not vote for themselves
	for _, ranking := range voteResp.Rankings {
		if ranking == a.ID {
			return nil, fmt.Errorf("agent %d voted for their own solution", a.ID)
		}
	}

	// Validate: ensure rankings are valid agent IDs (1 to Total, excluding self)
	for _, ranking := range voteResp.Rankings {
		if ranking < 1 || ranking > a.Total {
			return nil, fmt.Errorf("invalid agent ID in rankings: %d", ranking)
		}
	}

	return &types.Vote{
		VoterID:   a.ID,
		Rankings:  voteResp.Rankings,
		Reasoning: voteResp.Reasoning,
	}, nil
}

// extractJSON extracts a JSON object from text that may contain markdown or other content
func extractJSON(s string) string {
	// First, try to find JSON in a code block
	codeBlockPattern := regexp.MustCompile("```(?:json)?\\s*\\n?([\\s\\S]*?)\\n?```")
	if matches := codeBlockPattern.FindStringSubmatch(s); len(matches) > 1 {
		s = matches[1]
	}

	// Find the first { and last } to extract the JSON object
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")

	if start == -1 || end == -1 || end <= start {
		return ""
	}

	return s[start : end+1]
}
