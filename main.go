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
	filter        string
	filtering     bool
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

func (m *model) visibleItems() []DirItem {
	if m.filter == "" {
		return m.dirItems
	}
	var filtered []DirItem
	filterLower := strings.ToLower(m.filter)
	for _, item := range m.dirItems {
		if strings.Contains(strings.ToLower(item.name), filterLower) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (m *model) getHeaderView() string {
	width := m.viewport.Width
	bg := m.theme.primaryBackground

	titleStyle := styleRenderer.NewStyle().
		Bold(true).
		Foreground(m.theme.title).
		Background(bg).
		PaddingLeft(2)

	title := titleStyle.Render("ccd")
	titleWidth := lipgloss.Width(title)

	sepStyle := styleRenderer.NewStyle().
		Foreground(m.theme.border).
		Background(bg)

	sepLen := max(0, width-titleWidth-1)
	sep := sepStyle.Render(strings.Repeat("─", sepLen))

	space := styleRenderer.NewStyle().Background(bg).Render(" ")
	headerTop := title + space + sep

	pathParts := strings.Split(m.dir, "/")
	breadcrumbStyle := styleRenderer.NewStyle().
		Foreground(m.theme.muted).
		Background(bg).
		PaddingLeft(2).
		Width(width)

	var crumbs []string
	for _, part := range pathParts {
		if part == "" {
			crumbs = append(crumbs, "/")
			continue
		}
		crumbs = append(crumbs, part)
	}

	pathSep := " / "
	breadcrumb := breadcrumbStyle.Render(strings.Join(crumbs, pathSep))

	var filterLine string
	if m.filtering {
		filterStyle := styleRenderer.NewStyle().
			Foreground(m.theme.accent).
			Background(bg).
			PaddingLeft(2).
			PaddingBottom(1).
			Width(width)
		cursor := "▎"
		filterLine = filterStyle.Render(fmt.Sprintf("/%s%s", m.filter, cursor))
	}

	bottomBorder := sepStyle.Render(strings.Repeat("─", width))

	if m.filtering {
		return strings.Join([]string{headerTop, breadcrumb, filterLine, bottomBorder}, "\n")
	}

	return strings.Join([]string{headerTop, breadcrumb, bottomBorder}, "\n")
}

func (m *model) getFooterView() string {
	width := m.viewport.Width
	bg := m.theme.primaryBackground
	padStyle := styleRenderer.NewStyle().Background(bg)

	badgeStyle := styleRenderer.NewStyle().
		Foreground(m.theme.primaryContent).
		Background(m.theme.accent).
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1)

	badgeMutedStyle := styleRenderer.NewStyle().
		Foreground(m.theme.primaryContent).
		Background(m.theme.tertiaryBackground).
		Bold(true).
		PaddingLeft(1).
		PaddingRight(1)

	hintStyle := styleRenderer.NewStyle().
		Foreground(m.theme.muted).
		Background(bg).
		PaddingRight(2)

	hiddenStatus := "off"
	hiddenBadge := badgeMutedStyle.Render("h")
	if m.showHidden {
		hiddenStatus = "on"
		hiddenBadge = badgeStyle.Render("h")
	}
	hiddenHint := lipgloss.JoinHorizontal(lipgloss.Left,
		hiddenBadge,
		hintStyle.Render(fmt.Sprintf(" hidden:%s", hiddenStatus)),
	)

	filterBadge := badgeMutedStyle.Render("^f")
	if m.filtering {
		filterBadge = badgeStyle.Render("^f")
	}
	filterHint := lipgloss.JoinHorizontal(lipgloss.Left,
		filterBadge,
		hintStyle.Render(" filter"),
	)

	themeHint := lipgloss.JoinHorizontal(lipgloss.Left,
		badgeStyle.Render("t"),
		hintStyle.Render(" theme"),
	)

	quitHint := lipgloss.JoinHorizontal(lipgloss.Left,
		badgeMutedStyle.Render("q"),
		hintStyle.Render(" quit"),
	)

	cancelHint := lipgloss.JoinHorizontal(lipgloss.Left,
		badgeStyle.Render("^c"),
		hintStyle.Render(" cancel"),
	)

	borderStyle := styleRenderer.NewStyle().
		Foreground(m.theme.border).
		Background(bg)

	topBorder := borderStyle.Render(strings.Repeat("─", width))

	errorText := ""
	if m.error != nil {
		errorStyle := styleRenderer.NewStyle().
			Foreground(m.theme.error).
			Background(bg).
			Bold(true).
			PaddingLeft(2)
		errorRaw := errorStyle.Render(fmt.Sprintf("  %s", m.error.Error()))
		errWidth := lipgloss.Width(errorRaw)
		if errWidth < width {
			errorRaw += padStyle.Render(strings.Repeat(" ", width-errWidth))
		}
		errorText = errorRaw
	}

	controls := lipgloss.JoinHorizontal(lipgloss.Center, hiddenHint, filterHint, themeHint, quitHint, cancelHint)
	controlsWidth := lipgloss.Width(controls)
	if controlsWidth < width {
		controls += padStyle.Render(strings.Repeat(" ", width-controlsWidth))
	}

	parts := []string{topBorder, controls}
	if m.error != nil {
		parts = append(parts, errorText)
	} else {
		parts = append(parts, padStyle.Render(strings.Repeat(" ", width)))
	}

	return strings.Join(parts, "\n")
}

func (m *model) getContent() string {
	bg := m.theme.primaryBackground

	activeBarStyle := styleRenderer.NewStyle().
		Foreground(m.theme.accent).
		Bold(true)

	activeStyle := styleRenderer.NewStyle().
		Foreground(m.theme.selectedFg).
		Background(m.theme.selectedBg).
		Bold(true)

	defaultStyle := styleRenderer.NewStyle().
		Foreground(m.theme.primaryContent)

	items := m.visibleItems()

	if len(items) == 0 {
		emptyStyle := styleRenderer.NewStyle().
			Foreground(m.theme.muted).
			Background(bg).
			Faint(true).
			PaddingLeft(2).
			PaddingTop(1).
			Width(m.viewport.Width)

		if m.filter != "" {
			return emptyStyle.Render(fmt.Sprintf("  (no matches for \"%s\")", m.filter))
		}
		return emptyStyle.Render("  (no subdirectories)")
	}

	rowStyle := styleRenderer.NewStyle().
		Background(bg)

	tbl := table.New().
		Width(m.viewport.Width).
		BorderTop(false).
		BorderRight(false).
		BorderBottom(false).
		BorderLeft(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			return rowStyle
		})

	for i, v := range items {
		iconChar := " "

		var line string
		if i == m.active {
			bar := activeBarStyle.Render(" ")
			line = bar + activeStyle.Render(fmt.Sprintf("  %s %s ", iconChar, v.name))
		} else if v.isHidden {
			line = defaultStyle.Render(fmt.Sprintf("   %s %s ", iconChar, v.name))
		} else {
			line = defaultStyle.Render(fmt.Sprintf("   %s %s ", iconChar, v.name))
		}

		tbl.Row(line)
	}

	contentStyle := styleRenderer.NewStyle().
		Background(bg).
		Width(m.viewport.Width)
	return contentStyle.Render(tbl.Render())
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
		if m.filtering {
			switch msg.String() {
			case "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "esc":
				m.filter = ""
				m.filtering = false
				m.active = 0
				m.viewport.SetContent(m.getContent())
				m.viewport.SetYOffset(m.active - m.viewport.VisibleLineCount()/2)
			case "backspace":
				if len(m.filter) > 0 {
					m.filter = m.filter[:len(m.filter)-1]
					m.active = 0
					m.viewport.SetContent(m.getContent())
					m.viewport.SetYOffset(m.active - m.viewport.VisibleLineCount()/2)
				}
			case "up":
				if m.active > 0 {
					m.active -= 1
				}
				m.viewport.SetContent(m.getContent())
				m.viewport.SetYOffset(m.active - m.viewport.VisibleLineCount()/2)
			case "down":
				items := m.visibleItems()
				if m.active < len(items)-1 {
					m.active += 1
				}
				m.viewport.SetContent(m.getContent())
				m.viewport.SetYOffset(m.active - m.viewport.VisibleLineCount()/2)
			case "enter", "right":
				items := m.visibleItems()
				if m.active >= len(items) {
					break
				}
				m.activeHistory[m.dir] = m.active
				m.dir = PushPath(m.dir, items[m.active].name)
				m.dirItems, m.error = List(m.dir, m.showHidden)
				m.filter = ""
				m.filtering = false
				active, containsKey := m.activeHistory[m.dir]
				if containsKey {
					m.active = active
				} else {
					m.active = 0
				}
				m.viewport.SetContent(m.getContent())
				m.viewport.SetYOffset(m.active - m.viewport.VisibleLineCount()/2)
			default:
				if msg.Type == tea.KeyRunes {
					for _, r := range msg.Runes {
						m.filter += string(r)
					}
					m.active = 0
					m.viewport.SetContent(m.getContent())
					m.viewport.SetYOffset(m.active - m.viewport.VisibleLineCount()/2)
				}
			}
		} else {
			switch msg.String() {
			case "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "q":
				m.quitting = true
				m.chosenDir = m.dir
				return m, tea.Quit
			case "ctrl+f":
				m.filtering = true
				m.filter = ""
				m.viewport.SetContent(m.getContent())
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
				m.filter = ""
				m.filtering = false

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
				m.filter = ""
				m.filtering = false

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

				m.viewport.Style = styleRenderer.NewStyle().
					Background(m.theme.primaryBackground)
				m.viewport.SetContent(m.getContent())
			}
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

		m.viewport.Style = styleRenderer.NewStyle().
			Background(m.theme.primaryBackground)
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
