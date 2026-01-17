package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/humzahkiani/council/internal/storage"
	"github.com/humzahkiani/council/internal/types"
)

// SessionItem represents a session in the list
type SessionItem struct {
	Path    string
	Session *types.Session
}

func (i SessionItem) Title() string {
	if i.Session.IsTie {
		return fmt.Sprintf("TIE - %s", truncate(i.Session.Task, 50))
	}
	winner := "?"
	if i.Session.WinnerID != nil {
		winner = fmt.Sprintf("Agent %d", *i.Session.WinnerID)
	}
	return fmt.Sprintf("%s won - %s", winner, truncate(i.Session.Task, 40))
}

func (i SessionItem) Description() string {
	return fmt.Sprintf("%s | %d agents | %s",
		i.Session.CreatedAt.Format("2006-01-02 15:04"),
		i.Session.AgentCount,
		i.Session.Model,
	)
}

func (i SessionItem) FilterValue() string {
	return i.Session.Task
}

// ListModel is the model for the session list view
type ListModel struct {
	list     list.Model
	selected *SessionItem
	quitting bool
	err      error
}

// NewListModel creates a new list model with sessions from ~/.council/sessions/
func NewListModel() (ListModel, error) {
	store, err := storage.New()
	if err != nil {
		return ListModel{}, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ListModel{}, err
	}

	sessionsDir := filepath.Join(homeDir, ".council", "sessions")
	files, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// No sessions yet
			return ListModel{
				list: list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
			}, nil
		}
		return ListModel{}, err
	}

	var items []list.Item
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		path := filepath.Join(sessionsDir, file.Name())
		session, err := store.Load(path)
		if err != nil {
			continue // Skip invalid files
		}

		items = append(items, SessionItem{
			Path:    path,
			Session: session,
		})
	}

	// Sort by creation time, newest first
	sort.Slice(items, func(i, j int) bool {
		si := items[i].(SessionItem)
		sj := items[j].(SessionItem)
		return si.Session.CreatedAt.After(sj.Session.CreatedAt)
	})

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(primaryColor).
		BorderLeftForeground(primaryColor)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(mutedColor).
		BorderLeftForeground(primaryColor)

	l := list.New(items, delegate, 0, 0)
	l.Title = "Council Sessions"
	l.Styles.Title = titleStyle
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)

	return ListModel{list: l}, nil
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(SessionItem); ok {
				m.selected = &item
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	if len(m.list.Items()) == 0 {
		return lipgloss.NewStyle().Margin(2).Render(
			titleStyle.Render("Council Sessions") + "\n\n" +
				mutedTextStyle.Render("No sessions found.") + "\n\n" +
				mutedTextStyle.Render("Run a council session first:") + "\n" +
				contentStyle.Render("  council --save \"Your task here\"") + "\n\n" +
				helpStyle.Render("Press q to quit."),
		)
	}

	return lipgloss.NewStyle().Margin(1, 2).Render(m.list.View())
}

// Selected returns the selected session, if any
func (m ListModel) Selected() *SessionItem {
	return m.selected
}

// Quitting returns whether the user quit without selecting
func (m ListModel) Quitting() bool {
	return m.quitting
}
