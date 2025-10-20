package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
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
	ready         bool
	viewport      viewport.Model
	error         error
}

func initialModel() model {
	defaultDir, err := os.Getwd()

	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	defaultShowHidden := false
	dirItems, err := List(defaultDir, defaultShowHidden)

	return model{
		dir:           defaultDir,
		dirItems:      dirItems,
		activeHistory: map[string]int{},
		active:        0,
		quitting:      false,
		chosenDir:     "",
		showHidden:    defaultShowHidden,
		error:         err,
	}
}

func (m *model) getHeaderView() string {
	var currentDirStyle = styleRenderer.NewStyle().Width(m.viewport.Width).PaddingLeft(2).PaddingBottom(1)
	return currentDirStyle.Render(m.dir)
}

func (m *model) getFooterView() string {
	footerStyle := styleRenderer.NewStyle().Width(m.viewport.Width)

	hidden := "off"
	if m.showHidden {
		hidden = "on"
	}

	hiddenStyle := styleRenderer.NewStyle().PaddingLeft(2).Faint(true)
	hiddenText := hiddenStyle.Render(fmt.Sprintf("Hidden: %s", hidden))

	errorText := ""
	if m.error != nil {
		errorStyle := styleRenderer.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("1"))
		errorText = errorStyle.Render(m.error.Error())
	}

	return footerStyle.Render(lipgloss.JoinHorizontal(0, hiddenText, errorText))
}

func (m *model) getContent() string {
	s := ""

	var activeStyle = styleRenderer.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	var defaultStyle = styleRenderer.NewStyle()

	for i, v := range m.dirItems {
		var line string

		if i == m.active {
			line = activeStyle.Render(fmt.Sprintf("> %s", v.name))
		} else {
			line = defaultStyle.Render(fmt.Sprintf("  %s", v.name))
		}

		if i == 0 {
			s = line
		} else {
			s = lipgloss.JoinVertical(0, s, line)
		}
	}

	return s
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

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
			m.dirItems, m.error = List(m.dir, m.showHidden)
			m.active = 0

			m.viewport.SetContent(m.getContent())
		case "up":
			if m.active > 0 {
				m.active -= 1
			}

			m.viewport.SetContent(m.getContent())
		case "down":
			if m.active < len(m.dirItems)-1 {
				m.active += 1
			}

			m.viewport.SetContent(m.getContent())
		case "enter", "right":
			if m.active >= len(m.dirItems) {
				break
			}

			m.activeHistory[m.dir] = m.active
			m.dir = PushPath(m.dir, m.dirItems[m.active].name)
			m.dirItems, m.error = List(m.dir, m.showHidden)

			active, containsKey := m.activeHistory[m.dir]
			if containsKey {
				m.active = active
			} else {
				m.active = 0
			}

			m.viewport.SetContent(m.getContent())
			m.viewport.SetYOffset(m.active)
		case "backspace", "esc", "left":
			m.activeHistory[m.dir] = m.active
			m.dir = PopPath(m.dir)
			m.dirItems, m.error = List(m.dir, m.showHidden)

			active, containsKey := m.activeHistory[m.dir]
			if containsKey {
				m.active = active
			} else {
				m.active = 0
			}

			m.viewport.SetContent(m.getContent())
			m.viewport.SetYOffset(m.active)
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.getHeaderView())
		footerHeight := lipgloss.Height(m.getFooterView())

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-headerHeight-footerHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.getContent())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerHeight - footerHeight
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	if !m.ready {
		return "\n Initializing..."
	}

	return fmt.Sprintf("%s\n%s\n%s", m.getHeaderView(), m.viewport.View(), m.getFooterView())
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
