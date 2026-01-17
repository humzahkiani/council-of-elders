package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED") // Purple
	secondaryColor = lipgloss.Color("#10B981") // Green
	mutedColor     = lipgloss.Color("#6B7280") // Gray
	errorColor     = lipgloss.Color("#EF4444") // Red
	warningColor   = lipgloss.Color("#F59E0B") // Yellow

	// Tab styles
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Padding(0, 2)

	tabGapStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1, 2)

	activePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1, 2)

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	subHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(secondaryColor)

	// Content styles
	contentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	mutedTextStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Winner highlight
	winnerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(secondaryColor)

	// Agent labels
	agentLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 1)

	// Score styles
	scoreStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(mutedColor).
			Padding(0, 1)

	winnerScoreStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(secondaryColor).
				Padding(0, 1)

	// Help bar
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	helpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))

	// Title
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	// List item styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	// Divider
	dividerStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
)

// Helper function to create a horizontal divider
func divider(width int) string {
	return dividerStyle.Render(lipgloss.NewStyle().Width(width).Render(repeatChar("─", width)))
}

func repeatChar(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}
