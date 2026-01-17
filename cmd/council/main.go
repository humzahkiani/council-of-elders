package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/humzahkiani/council/internal/council"
	"github.com/humzahkiani/council/internal/storage"
	"github.com/humzahkiani/council/internal/tui"
	"github.com/humzahkiani/council/internal/types"
)

var (
	agentCount int
	rounds     int
	save       bool
	outputPath string
	verbose    bool
	model      string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "council",
		Short: "Council of Elders - Multi-agent deliberation with ranked-choice voting",
		Long: `Council presents a task to multiple Claude agents, facilitates discussion
among them, and uses ranked-choice voting to determine the best solution.

Examples:
  council run "Write a function to check if a number is prime"
  council run --agents 5 --rounds 2 "Design a REST API for a blog"
  council view                    # List all sessions
  council view <session-id>       # View a specific session`,
	}

	// Run subcommand
	runCmd := &cobra.Command{
		Use:   "run [task]",
		Short: "Run a council deliberation on a task",
		Long: `Run a council of AI agents to deliberate on a task.

Examples:
  council run "Write a function to check if a number is prime"
  council run --agents 5 --rounds 2 --save "Design a REST API for a blog"`,
		Args:    cobra.ExactArgs(1),
		PreRunE: validateRun,
		RunE:    runCouncil,
	}

	runCmd.Flags().IntVarP(&agentCount, "agents", "a", 3, "Number of agents (minimum 3)")
	runCmd.Flags().IntVarP(&rounds, "rounds", "r", 1, "Number of discussion rounds")
	runCmd.Flags().BoolVarP(&save, "save", "s", false, "Save session to ~/.council/sessions/")
	runCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Save session to specific file path")
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Print detailed output during execution")
	runCmd.Flags().StringVarP(&model, "model", "m", "claude-sonnet-4-20250514", "Claude model to use")

	// View subcommand
	viewCmd := &cobra.Command{
		Use:   "view [session-id]",
		Short: "View a council session in the TUI",
		Long: `View a council session in an interactive TUI.

Without arguments, shows a list of all saved sessions.
With a session ID or file path, opens that specific session.

Examples:
  council view                           # List all sessions
  council view 2024-01-16_143022_abc123  # View by ID
  council view ./my-session.json         # View by file path`,
		Args: cobra.MaximumNArgs(1),
		RunE: viewSession,
	}

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(viewCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func validateRun(cmd *cobra.Command, args []string) error {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	if agentCount < 3 {
		return fmt.Errorf("minimum 3 agents required (got %d)", agentCount)
	}

	if rounds < 1 {
		return fmt.Errorf("minimum 1 discussion round required (got %d)", rounds)
	}

	return nil
}

func runCouncil(cmd *cobra.Command, args []string) error {
	config := &types.Config{
		AgentCount: agentCount,
		Rounds:     rounds,
		Save:       save,
		OutputPath: outputPath,
		Verbose:    verbose,
		Model:      model,
		Task:       args[0],
	}

	c, err := council.New(config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nInterrupted. Shutting down...")
		cancel()
	}()

	if err := c.Run(ctx); err != nil {
		return err
	}

	c.Output()
	return nil
}

func viewSession(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		// Show session list
		return showSessionList()
	}

	// Load specific session
	sessionPath := args[0]

	// Check if it's a file path or a session ID
	if !strings.HasSuffix(sessionPath, ".json") {
		// It's a session ID, look in ~/.council/sessions/
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		sessionsDir := filepath.Join(homeDir, ".council", "sessions")
		files, err := os.ReadDir(sessionsDir)
		if err != nil {
			return fmt.Errorf("failed to read sessions directory: %w", err)
		}

		// Find matching session
		found := false
		for _, file := range files {
			if strings.Contains(file.Name(), sessionPath) {
				sessionPath = filepath.Join(sessionsDir, file.Name())
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("session not found: %s", args[0])
		}
	}

	return showSession(sessionPath)
}

func showSessionList() error {
	listModel, err := tui.NewListModel()
	if err != nil {
		return fmt.Errorf("failed to load sessions: %w", err)
	}

	p := tea.NewProgram(listModel, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Check if user selected a session
	if lm, ok := finalModel.(tui.ListModel); ok {
		if selected := lm.Selected(); selected != nil {
			// Open the selected session
			return showSession(selected.Path)
		}
	}

	return nil
}

func showSession(path string) error {
	store, err := storage.New()
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	session, err := store.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	model := tui.NewModel(session)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
