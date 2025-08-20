package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var styleRenderer = lipgloss.NewRenderer(os.Stderr)

type model struct {
	dir           string
	dirItems      []DirItem
	activeHistory map[string]int
	active        int
	quitting      bool
	chosenDir     string
	showHidden    bool
}

func initialModel() model {
	defaultDir, err := os.Getwd()

	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	defaultShowHidden := false

	return model{
		dir:           defaultDir,
		dirItems:      List(defaultDir, defaultShowHidden),
		activeHistory: map[string]int{},
		active:        0,
		quitting:      false,
		chosenDir:     "",
		showHidden:    defaultShowHidden,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			m.quitting = true
			m.chosenDir = m.dir
			return m, tea.Quit
		case "h":
			m.showHidden = !m.showHidden
			m.dirItems = List(m.dir, m.showHidden)
			m.active = 0
		case "up":
			if m.active > 0 {
				m.active -= 1
			} else {
				m.active = len(m.dirItems) - 1
			}
		case "down":
			if m.active < len(m.dirItems)-1 {
				m.active += 1
			} else {
				m.active = 0
			}
		case "enter":
			m.activeHistory[m.dir] = m.active
			m.dir = PushPath(m.dir, m.dirItems[m.active].name)
			m.dirItems = List(m.dir, m.showHidden)

			active, containsKey := m.activeHistory[m.dir]
			if containsKey {
				m.active = active
			} else {
				m.active = 0
			}
		case "backspace", "esc":
			m.activeHistory[m.dir] = m.active
			m.dir = PopPath(m.dir)
			m.dirItems = List(m.dir, m.showHidden)

			active, containsKey := m.activeHistory[m.dir]
			if containsKey {
				m.active = active
			} else {
				m.active = 0
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var currentDirStyle = styleRenderer.NewStyle().PaddingLeft(2).BorderStyle(lipgloss.MarkdownBorder()).BorderBottom(true)

	s := currentDirStyle.Render(m.dir)

	var activeStyle = styleRenderer.NewStyle().Foreground(lipgloss.Color("#FF0066"))
	var defaultStyle = styleRenderer.NewStyle()

	for i, v := range m.dirItems {
		if i == m.active {
			s = lipgloss.JoinVertical(0, s, activeStyle.Render(fmt.Sprintf("> %s", v.name)))
		} else {
			s = lipgloss.JoinVertical(0, s, defaultStyle.Render(fmt.Sprintf("  %s", v.name)))
		}
	}

	return s
}

func main() {
	fm, err := tea.NewProgram(initialModel(), tea.WithOutput(os.Stderr), tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	m := fm.(model)

	if m.chosenDir != "" {
		fmt.Print(m.dir)
	}

	os.Exit(0)
}
