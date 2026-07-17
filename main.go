package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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
	theme         Theme
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
		theme:         darkTheme,
	}
}

func (m *model) getHeaderView() string {
	width := m.viewport.Width

	titleStyle := styleRenderer.NewStyle().
		Bold(true).
		Foreground(m.theme.title).
		PaddingLeft(2)

	title := titleStyle.Render("ccd")

	sepStyle := styleRenderer.NewStyle().
		Foreground(m.theme.border).
		Faint(true)

	sep := sepStyle.Render(strings.Repeat("─", max(0, width-10)))

	headerTop := lipgloss.JoinHorizontal(lipgloss.Center, title, " ", sep)

	pathParts := strings.Split(m.dir, "/")
	breadcrumbStyle := styleRenderer.NewStyle().
		Foreground(m.theme.muted).
		PaddingLeft(2).
		PaddingTop(1).
		PaddingBottom(1)

	var crumbs []string
	for _, part := range pathParts {
		if part == "" {
			crumbs = append(crumbs, "/")
			continue
		}
		crumbs = append(crumbs, part)
	}

	pathSep := sepStyle.Render(" / ")
	breadcrumb := breadcrumbStyle.Render(strings.Join(crumbs, pathSep))

	bottomBorder := styleRenderer.NewStyle().
		Foreground(m.theme.border).
		Render(strings.Repeat("─", max(0, width-2)))

	return lipgloss.JoinVertical(lipgloss.Left, headerTop, breadcrumb, bottomBorder)
}

func (m *model) getFooterView() string {
	width := m.viewport.Width

	badgeStyle := styleRenderer.NewStyle().
		Foreground(m.theme.primaryBackground).
		Background(m.theme.accent).
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1)

	badgeMutedStyle := styleRenderer.NewStyle().
		Foreground(m.theme.primaryBackground).
		Background(m.theme.muted).
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1)

	hintStyle := styleRenderer.NewStyle().
		Foreground(m.theme.muted).
		PaddingRight(2)

	hiddenStatus := "off"
	hiddenBadge := badgeMutedStyle.Render("h")
	if m.showHidden {
		hiddenStatus = "on"
		hiddenBadge = badgeStyle.Render("h")
	}
	hiddenHint := hintStyle.Render(fmt.Sprintf("%s hidden:%s", hiddenBadge, hiddenStatus))

	themeBadge := badgeStyle.Render("t")
	themeHint := hintStyle.Render(fmt.Sprintf("%s theme", themeBadge))

	quitBadge := badgeMutedStyle.Render("q")
	quitHint := hintStyle.Render(fmt.Sprintf("%s quit", quitBadge))

	cancelBadge := badgeStyle.Render("^c")
	cancelHint := hintStyle.Render(fmt.Sprintf("%s cancel", cancelBadge))

	topBorder := styleRenderer.NewStyle().
		Foreground(m.theme.border).
		Render(strings.Repeat("─", max(0, width-2)))

	errorText := ""
	if m.error != nil {
		errorStyle := styleRenderer.NewStyle().
			Foreground(m.theme.error).
			Bold(true).
			PaddingLeft(2)
		errorText = errorStyle.Render(fmt.Sprintf("  %s", m.error.Error()))
	}

	controls := lipgloss.JoinHorizontal(lipgloss.Center, hiddenHint, themeHint, quitHint, cancelHint)
	footer := lipgloss.JoinVertical(lipgloss.Left, topBorder, controls, errorText)

	return footer
}

func (m *model) getContent() string {
	activeBarStyle := styleRenderer.NewStyle().
		Foreground(m.theme.accent).
		Bold(true)

	activeStyle := styleRenderer.NewStyle().
		Foreground(m.theme.selectedFg).
		Background(m.theme.selectedBg).
		Bold(true)

	defaultStyle := styleRenderer.NewStyle().
		Foreground(m.theme.primaryContent)

	hiddenStyle := styleRenderer.NewStyle().
		Foreground(m.theme.muted)

	emptyStyle := styleRenderer.NewStyle().
		Foreground(m.theme.muted).
		Faint(true).
		PaddingLeft(2).
		PaddingTop(1)

	if len(m.dirItems) == 0 {
		return emptyStyle.Render("  (no subdirectories)")
	}

	rowStyle := styleRenderer.NewStyle().
		Background(m.theme.primaryBackground)

	tbl := table.New().
		Width(m.viewport.Width).
		BorderTop(false).
		BorderRight(false).
		BorderBottom(false).
		BorderLeft(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			return rowStyle
		})

	for i, v := range m.dirItems {
		iconChar := " "
		if v.isHidden {
			iconChar = " "
		}

		var line string
		if i == m.active {
			bar := activeBarStyle.Render(" ")
			line = bar + activeStyle.Render(fmt.Sprintf(" %s %s ", iconChar, v.name))
		} else if v.isHidden {
			line = hiddenStyle.Render(fmt.Sprintf("   %s %s ", iconChar, v.name))
		} else {
			line = defaultStyle.Render(fmt.Sprintf("   %s %s ", iconChar, v.name))
		}

		tbl.Row(line)
	}

	return tbl.Render()
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
			m.viewport.SetYOffset(m.active - m.viewport.VisibleLineCount()/2)
		case "up":
			if m.active > 0 {
				m.active -= 1
			}

			m.viewport.SetContent(m.getContent())
			m.viewport.SetYOffset(m.active - m.viewport.VisibleLineCount()/2)
		case "down":
			if m.active < len(m.dirItems)-1 {
				m.active += 1
			}

			m.viewport.SetContent(m.getContent())
			m.viewport.SetYOffset(m.active - m.viewport.VisibleLineCount()/2)
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
			m.viewport.SetYOffset(m.active - m.viewport.VisibleLineCount()/2)
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
			m.viewport.SetYOffset(m.active - m.viewport.VisibleLineCount()/2)
		case "t":
			if m.theme == darkTheme {
				m.theme = lightTheme
			} else {
				m.theme = darkTheme
			}

			m.viewport.SetContent(m.getContent())
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
	appStyle := styleRenderer.NewStyle().Background(m.theme.primaryBackground)

	if m.quitting {
		return ""
	}

	if !m.ready {
		return "\n Initializing..."
	}

	return appStyle.Render(fmt.Sprintf("%s\n%s\n%s", m.getHeaderView(), m.viewport.View(), m.getFooterView()))
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
