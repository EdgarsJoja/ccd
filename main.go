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
	dir            string
	dirItems       []DirItem
	activeHistory  map[string]int
	active         int
	quitting       bool
	chosenDir      string
	showHidden     bool
	ready          bool
	viewport       viewport.Model
	error          error
	theme          Theme
	filter         string
	filtering      bool
	shortcuts      []Shortcut
	showShortcuts  bool
	shortcutActive int
	naming         bool
	shortcutName   string
}

func initialModel() model {
	defaultDir, err := os.Getwd()

	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	defaultShowHidden := false
	dirItems, err := List(defaultDir, defaultShowHidden)

	shortcuts, shortcutsErr := LoadShortcuts()
	if err == nil {
		err = shortcutsErr
	}

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
		shortcuts:     shortcuts,
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
	pathLabel := strings.Join(crumbs, pathSep)
	if m.showShortcuts {
		pathLabel += "  ·  shortcuts"
	}
	breadcrumb := breadcrumbStyle.Render(pathLabel)

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

	shortcutBadge := badgeMutedStyle.Render("s")
	shortcutLabel := " shortcuts"
	if m.showShortcuts {
		shortcutBadge = badgeStyle.Render("s")
		shortcutLabel = " back"
	}
	shortcutHint := lipgloss.JoinHorizontal(lipgloss.Left,
		shortcutBadge,
		hintStyle.Render(shortcutLabel),
	)

	addHint := lipgloss.JoinHorizontal(lipgloss.Left,
		badgeMutedStyle.Render("a"),
		hintStyle.Render(" add"),
	)

	deleteHint := lipgloss.JoinHorizontal(lipgloss.Left,
		badgeStyle.Render("d"),
		hintStyle.Render(" delete"),
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

	var controlHints []string
	if m.naming {
		enterHint := lipgloss.JoinHorizontal(lipgloss.Left,
			badgeStyle.Render("enter"),
			hintStyle.Render(" save"),
		)
		escHint := lipgloss.JoinHorizontal(lipgloss.Left,
			badgeStyle.Render("esc"),
			hintStyle.Render(" cancel"),
		)
		controlHints = []string{enterHint, escHint, cancelHint}
	} else {
		controlHints = []string{hiddenHint, filterHint, shortcutHint, addHint}
		if m.showShortcuts {
			controlHints = append(controlHints, deleteHint)
		}
		controlHints = append(controlHints, themeHint, quitHint, cancelHint)
	}
	controls := lipgloss.JoinHorizontal(lipgloss.Center, controlHints...)
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
	var content string
	if m.showShortcuts {
		content = m.getShortcutsContent()
	} else {
		content = m.getDirContent()
	}

	if m.naming {
		content = m.getShortcutPrompt() + "\n" + content
	}

	return content
}

func (m *model) getShortcutPrompt() string {
	promptStyle := styleRenderer.NewStyle().
		Foreground(m.theme.accent).
		Background(m.theme.primaryBackground).
		Bold(true).
		PaddingLeft(2).
		Width(m.viewport.Width)

	prompt := fmt.Sprintf("Add shortcut for %s as: %s▎  (enter save · esc cancel)", m.dir, m.shortcutName)

	return promptStyle.Render(prompt)
}

func (m *model) getShortcutsContent() string {
	bg := m.theme.primaryBackground

	activeBarStyle := styleRenderer.NewStyle().
		Foreground(m.theme.accent).
		Bold(true)

	activeNameStyle := styleRenderer.NewStyle().
		Foreground(m.theme.selectedFg).
		Background(m.theme.selectedBg).
		Bold(true)

	activePathStyle := styleRenderer.NewStyle().
		Foreground(m.theme.selectedFg).
		Background(m.theme.selectedBg)

	nameStyle := styleRenderer.NewStyle().
		Foreground(m.theme.primaryContent)

	pathStyle := styleRenderer.NewStyle().
		Foreground(m.theme.muted)

	if len(m.shortcuts) == 0 {
		emptyStyle := styleRenderer.NewStyle().
			Foreground(m.theme.muted).
			Background(bg).
			Faint(true).
			PaddingLeft(2).
			PaddingTop(1).
			Width(m.viewport.Width)

		return emptyStyle.Render("  (no shortcuts) — press a to add the current directory")
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

	for i, shortcut := range m.shortcuts {
		if i == m.shortcutActive {
			line := activeBarStyle.Render(" ") +
				activeNameStyle.Render(fmt.Sprintf("  %s ", shortcut.Name)) +
				activePathStyle.Render(shortcut.Path+" ")
			tbl.Row(line)
			continue
		}

		line := nameStyle.Render(fmt.Sprintf("   %s ", shortcut.Name)) +
			pathStyle.Render("  "+shortcut.Path+" ")
		tbl.Row(line)
	}

	contentStyle := styleRenderer.NewStyle().
		Background(bg).
		Width(m.viewport.Width)

	return contentStyle.Render(tbl.Render())
}

func (m *model) getDirContent() string {
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

func (m *model) activeIndex() int {
	if m.showShortcuts {
		return m.shortcutActive
	}

	return m.active
}

func (m *model) refreshView() {
	m.viewport.SetContent(m.getContent())
	if m.naming {
		m.viewport.SetYOffset(0)
		return
	}
	m.viewport.SetYOffset(m.activeIndex() - m.viewport.VisibleLineCount()/2)
}

func (m *model) startNaming() {
	m.naming = true
	m.shortcutName = DefaultShortcutName(m.dir)
	m.filter = ""
	m.filtering = false
	m.refreshView()
}

func (m *model) saveShortcut() {
	updated := AddShortcut(m.shortcuts, m.shortcutName, m.dir)
	if err := SaveShortcuts(updated); err != nil {
		m.error = err
		return
	}

	m.shortcuts = updated
	m.error = nil
	m.naming = false
	m.shortcutName = ""
}

func (m *model) removeShortcut() {
	if m.shortcutActive < 0 || m.shortcutActive >= len(m.shortcuts) {
		return
	}

	updated := RemoveShortcutAt(m.shortcuts, m.shortcutActive)
	if err := SaveShortcuts(updated); err != nil {
		m.error = err
		return
	}

	m.shortcuts = updated
	m.error = nil
	if m.shortcutActive >= len(m.shortcuts) {
		m.shortcutActive = len(m.shortcuts) - 1
	}
	if m.shortcutActive < 0 {
		m.shortcutActive = 0
	}
}

func (m *model) jumpToShortcut() {
	if m.shortcutActive < 0 || m.shortcutActive >= len(m.shortcuts) {
		return
	}

	shortcut := m.shortcuts[m.shortcutActive]
	info, err := os.Stat(shortcut.Path)
	if err != nil {
		m.error = fmt.Errorf("cannot open shortcut %q: %w", shortcut.Name, err)
		return
	}
	if !info.IsDir() {
		m.error = fmt.Errorf("shortcut %q is not a directory: %s", shortcut.Name, shortcut.Path)
		return
	}

	m.activeHistory[m.dir] = m.active
	m.dir = shortcut.Path
	m.dirItems, m.error = List(m.dir, m.showHidden)
	m.filter = ""
	m.filtering = false
	m.showShortcuts = false

	if active, ok := m.activeHistory[m.dir]; ok {
		m.active = active
	} else {
		m.active = 0
	}
}

func (m *model) toggleTheme() {
	if m.theme == darkTheme {
		m.theme = lightTheme
	} else {
		m.theme = darkTheme
	}

	m.viewport.Style = styleRenderer.NewStyle().
		Background(m.theme.primaryBackground)
	m.refreshView()
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
		if m.naming {
			switch msg.String() {
			case "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "esc":
				m.naming = false
				m.shortcutName = ""
				m.refreshView()
			case "enter":
				m.saveShortcut()
				m.refreshView()
			case "backspace":
				runes := []rune(m.shortcutName)
				if len(runes) > 0 {
					m.shortcutName = string(runes[:len(runes)-1])
				}
				m.refreshView()
			case "ctrl+u":
				m.shortcutName = ""
				m.refreshView()
			default:
				if msg.Type == tea.KeyRunes {
					for _, r := range msg.Runes {
						m.shortcutName += string(r)
					}
					m.refreshView()
				}
			}
		} else if m.filtering {
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
		} else if m.showShortcuts {
			switch msg.String() {
			case "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "q":
				m.quitting = true
				m.chosenDir = m.dir
				return m, tea.Quit
			case "s", "esc", "left", "backspace":
				m.showShortcuts = false
				m.shortcutActive = 0
				m.refreshView()
			case "a":
				m.startNaming()
			case "up":
				if m.shortcutActive > 0 {
					m.shortcutActive -= 1
				}
				m.refreshView()
			case "down":
				if m.shortcutActive < len(m.shortcuts)-1 {
					m.shortcutActive += 1
				}
				m.refreshView()
			case "enter", "right":
				m.jumpToShortcut()
				m.refreshView()
			case "d", "delete":
				m.removeShortcut()
				m.refreshView()
			case "t":
				m.toggleTheme()
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
			case "s":
				m.showShortcuts = true
				m.shortcutActive = 0
				m.filter = ""
				m.filtering = false
				m.refreshView()
			case "a":
				m.startNaming()
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
				m.toggleTheme()
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

	if _, isKey := msg.(tea.KeyMsg); !isKey || (!m.naming && !m.showShortcuts) {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

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
