package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/humzahkiani/council/internal/types"
)

// Tab represents a view tab
type Tab int

const (
	TabSolutions Tab = iota
	TabDiscussion
	TabVotes
	TabResults
)

func (t Tab) String() string {
	switch t {
	case TabSolutions:
		return "Solutions"
	case TabDiscussion:
		return "Discussion"
	case TabVotes:
		return "Votes"
	case TabResults:
		return "Results"
	default:
		return ""
	}
}

// Model is the main TUI model
type Model struct {
	session       *types.Session
	activeTab     Tab
	viewport      viewport.Model
	width         int
	height        int
	ready         bool
	selectedAgent int // For solutions tab, which agent's solution to highlight
}

// NewModel creates a new TUI model for viewing a session
func NewModel(session *types.Session) Model {
	return Model{
		session:       session,
		activeTab:     TabSolutions,
		selectedAgent: 1,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab", "l", "right":
			m.activeTab = Tab((int(m.activeTab) + 1) % 4)
			m.updateViewport()
		case "shift+tab", "h", "left":
			m.activeTab = Tab((int(m.activeTab) + 3) % 4) // +3 is same as -1 mod 4
			m.updateViewport()
		case "1":
			m.activeTab = TabSolutions
			m.updateViewport()
		case "2":
			m.activeTab = TabDiscussion
			m.updateViewport()
		case "3":
			m.activeTab = TabVotes
			m.updateViewport()
		case "4":
			m.activeTab = TabResults
			m.updateViewport()
		case "j", "down":
			if m.activeTab == TabSolutions {
				if m.selectedAgent < m.session.AgentCount {
					m.selectedAgent++
					m.updateViewport()
				}
			} else {
				m.viewport, cmd = m.viewport.Update(msg)
			}
		case "k", "up":
			if m.activeTab == TabSolutions {
				if m.selectedAgent > 1 {
					m.selectedAgent--
					m.updateViewport()
				}
			} else {
				m.viewport, cmd = m.viewport.Update(msg)
			}
		default:
			m.viewport, cmd = m.viewport.Update(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 6 // Title + tabs + padding
		footerHeight := 3 // Help text

		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, msg.Height-headerHeight-footerHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - headerHeight - footerHeight
		}
		m.updateViewport()
	}

	return m, cmd
}

// updateViewport updates the viewport content based on active tab
func (m *Model) updateViewport() {
	var content string
	switch m.activeTab {
	case TabSolutions:
		content = m.renderSolutions()
	case TabDiscussion:
		content = m.renderDiscussion()
	case TabVotes:
		content = m.renderVotes()
	case TabResults:
		content = m.renderResults()
	}
	m.viewport.SetContent(content)
}

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Title
	title := titleStyle.Render("Council of Elders")
	task := mutedTextStyle.Render(fmt.Sprintf("Task: %s", truncate(m.session.Task, m.width-10)))

	// Tabs
	tabs := m.renderTabs()

	// Content
	content := m.viewport.View()

	// Help
	help := m.renderHelp()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		task,
		"",
		tabs,
		content,
		help,
	)
}

// renderTabs renders the tab bar
func (m Model) renderTabs() string {
	var tabs []string
	allTabs := []Tab{TabSolutions, TabDiscussion, TabVotes, TabResults}

	for _, t := range allTabs {
		style := inactiveTabStyle
		if t == m.activeTab {
			style = activeTabStyle
		}
		tabs = append(tabs, style.Render(t.String()))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

// renderSolutions renders the solutions view
func (m Model) renderSolutions() string {
	var sb strings.Builder

	for _, sol := range m.session.Solutions {
		isWinner := m.session.WinnerID != nil && *m.session.WinnerID == sol.AgentID
		isSelected := sol.AgentID == m.selectedAgent

		// Agent header
		label := fmt.Sprintf("Agent %d", sol.AgentID)
		if isWinner {
			label += " ★ WINNER"
		}

		var headerStyleToUse lipgloss.Style
		if isSelected {
			headerStyleToUse = winnerStyle
		} else {
			headerStyleToUse = subHeaderStyle
		}

		sb.WriteString(headerStyleToUse.Render(label))
		sb.WriteString("\n")

		// Score
		score := m.session.Scores[sol.AgentID]
		scoreText := fmt.Sprintf("%d points", score)
		if isWinner {
			sb.WriteString(winnerScoreStyle.Render(scoreText))
		} else {
			sb.WriteString(scoreStyle.Render(scoreText))
		}
		sb.WriteString("\n\n")

		// Solution content
		sb.WriteString(contentStyle.Render(sol.Content))
		sb.WriteString("\n\n")
		sb.WriteString(divider(m.width - 8))
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// renderDiscussion renders the discussion view
func (m Model) renderDiscussion() string {
	if len(m.session.Critiques) == 0 {
		return mutedTextStyle.Render("No discussion recorded.")
	}

	var sb strings.Builder

	currentRound := 0
	for _, crit := range m.session.Critiques {
		if crit.Round != currentRound {
			currentRound = crit.Round
			sb.WriteString(headerStyle.Render(fmt.Sprintf("Round %d", currentRound)))
			sb.WriteString("\n\n")
		}

		sb.WriteString(subHeaderStyle.Render(fmt.Sprintf("Agent %d's Critique", crit.AgentID)))
		sb.WriteString("\n\n")
		sb.WriteString(contentStyle.Render(crit.Content))
		sb.WriteString("\n\n")
		sb.WriteString(divider(m.width - 8))
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// renderVotes renders the votes view
func (m Model) renderVotes() string {
	if len(m.session.Votes) == 0 {
		return mutedTextStyle.Render("No votes recorded.")
	}

	var sb strings.Builder

	sb.WriteString(headerStyle.Render("Vote Breakdown"))
	sb.WriteString("\n\n")

	for _, vote := range m.session.Votes {
		sb.WriteString(subHeaderStyle.Render(fmt.Sprintf("Agent %d's Vote", vote.VoterID)))
		sb.WriteString("\n")

		// Rankings
		sb.WriteString(mutedTextStyle.Render("Rankings: "))
		for i, agentID := range vote.Rankings {
			points := m.session.AgentCount - 1 - i
			if i > 0 {
				sb.WriteString(" → ")
			}
			rankText := fmt.Sprintf("Agent %d (%dpts)", agentID, points)
			if m.session.WinnerID != nil && *m.session.WinnerID == agentID {
				sb.WriteString(winnerStyle.Render(rankText))
			} else {
				sb.WriteString(contentStyle.Render(rankText))
			}
		}
		sb.WriteString("\n\n")

		// Reasoning
		if vote.Reasoning != "" {
			sb.WriteString(mutedTextStyle.Render("Reasoning: "))
			sb.WriteString(contentStyle.Render(vote.Reasoning))
			sb.WriteString("\n")
		}

		sb.WriteString("\n")
		sb.WriteString(divider(m.width - 8))
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// renderResults renders the results summary
func (m Model) renderResults() string {
	var sb strings.Builder

	sb.WriteString(headerStyle.Render("Final Results"))
	sb.WriteString("\n\n")

	// Scores table
	sb.WriteString(subHeaderStyle.Render("Scores"))
	sb.WriteString("\n\n")

	for i := 1; i <= m.session.AgentCount; i++ {
		score := m.session.Scores[i]
		isWinner := m.session.WinnerID != nil && *m.session.WinnerID == i
		isTied := m.session.IsTie && contains(m.session.TiedAgents, i)

		line := fmt.Sprintf("Agent %d: %d points", i, score)
		if isWinner {
			line += " ★ WINNER"
			sb.WriteString(winnerStyle.Render(line))
		} else if isTied {
			line += " (TIED)"
			sb.WriteString(warningStyle.Render(line))
		} else {
			sb.WriteString(contentStyle.Render(line))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(divider(m.width - 8))
	sb.WriteString("\n\n")

	// Winner announcement
	if m.session.IsTie {
		sb.WriteString(warningStyle.Render(fmt.Sprintf("TIE between Agents %v", m.session.TiedAgents)))
		sb.WriteString("\n\n")
		sb.WriteString(mutedTextStyle.Render("No single winner - review solutions to decide."))
	} else if m.session.WinnerID != nil {
		sb.WriteString(winnerStyle.Render(fmt.Sprintf("Winner: Agent %d", *m.session.WinnerID)))
		sb.WriteString("\n\n")

		// Show winning solution
		for _, sol := range m.session.Solutions {
			if sol.AgentID == *m.session.WinnerID {
				sb.WriteString(subHeaderStyle.Render("Winning Solution"))
				sb.WriteString("\n\n")
				sb.WriteString(contentStyle.Render(sol.Content))
				break
			}
		}
	}

	// Session metadata
	sb.WriteString("\n\n")
	sb.WriteString(divider(m.width - 8))
	sb.WriteString("\n\n")
	sb.WriteString(subHeaderStyle.Render("Session Info"))
	sb.WriteString("\n\n")
	sb.WriteString(mutedTextStyle.Render(fmt.Sprintf("ID: %s", m.session.ID)))
	sb.WriteString("\n")
	sb.WriteString(mutedTextStyle.Render(fmt.Sprintf("Model: %s", m.session.Model)))
	sb.WriteString("\n")
	sb.WriteString(mutedTextStyle.Render(fmt.Sprintf("Agents: %d", m.session.AgentCount)))
	sb.WriteString("\n")
	sb.WriteString(mutedTextStyle.Render(fmt.Sprintf("Rounds: %d", m.session.Rounds)))
	sb.WriteString("\n")
	sb.WriteString(mutedTextStyle.Render(fmt.Sprintf("Created: %s", m.session.CreatedAt.Format("2006-01-02 15:04:05"))))
	sb.WriteString("\n")
	sb.WriteString(mutedTextStyle.Render(fmt.Sprintf("Completed: %s", m.session.CompletedAt.Format("2006-01-02 15:04:05"))))

	return sb.String()
}

// renderHelp renders the help bar
func (m Model) renderHelp() string {
	keys := []string{
		helpKeyStyle.Render("1-4") + " tabs",
		helpKeyStyle.Render("←/→") + " switch",
		helpKeyStyle.Render("↑/↓") + " scroll",
		helpKeyStyle.Render("q") + " quit",
	}
	return helpStyle.Render(strings.Join(keys, "  │  "))
}

// Helper functions

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func contains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

var warningStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(warningColor)
