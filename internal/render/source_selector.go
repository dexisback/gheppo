package render

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dexisback/gheppo/internal/config"
)

// SourceOption represents a selectable data source.
type SourceOption struct {
	ID          string
	Title       string
	Description string
}

// SourceSelectorModel manages the interactive source selection state.
type SourceSelectorModel struct {
	options        []SourceOption
	cursor         int
	SelectedSource string
	Canceled       bool
	quitting       bool
	width          int
	height         int
	mode           ColorMode
	theme          config.Theme
}

// NewSourceSelectorModel initializes the source selector model with gruvbox styling.
func NewSourceSelectorModel() SourceSelectorModel {
	gruvboxTheme, _ := config.GetTheme("gruvbox")
	if gruvboxTheme.Name == "" {
		gruvboxTheme = config.DefaultTheme()
	}

	mode := DetectColorMode()
	termWidth := GetTerminalWidth()

	options := []SourceOption{
		{
			ID:          config.SourceGitHub,
			Title:       "GitHub",
			Description: "Your daily git commit graph & open-source contributions",
		},
		{
			ID:          config.SourceLeetCode,
			Title:       "LeetCode",
			Description: "Problem-solving submission calendar & competitive stats",
		},
	}

	currentSource := config.GetSource()
	cursor := 0
	for i, opt := range options {
		if strings.EqualFold(opt.ID, currentSource) {
			cursor = i
			break
		}
	}

	return SourceSelectorModel{
		options:  options,
		cursor:   cursor,
		width:    termWidth,
		height:   20,
		mode:     mode,
		theme:    gruvboxTheme,
	}
}

// Init starts the Bubble Tea lifecycle.
func (m SourceSelectorModel) Init() tea.Cmd {
	return nil
}

// Update handles navigation, selection, and cancellation.
func (m SourceSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if len(m.options) > 0 {
				m.cursor = (m.cursor - 1 + len(m.options)) % len(m.options)
			}
			return m, nil

		case "down", "j":
			if len(m.options) > 0 {
				m.cursor = (m.cursor + 1) % len(m.options)
			}
			return m, nil

		case "enter":
			if len(m.options) > 0 {
				m.SelectedSource = m.options[m.cursor].ID
			}
			m.quitting = true
			return m, tea.Quit

		case "esc", "q", "ctrl+c":
			m.Canceled = true
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View builds the compact source selector UI styled with the Gruvbox palette.
func (m SourceSelectorModel) View() string {
	if m.quitting {
		return ""
	}

	theme := ResolveTheme(m.theme, m.mode)
	cfg := m.theme

	totalWidth := 56

	// === 1. WORDMARK ===
	wLines, _ := RenderWordmarkWithStateAndWidth(theme, WordmarkState{Progress: 1.0}, 46)

	// === 2. SUBTITLE & INSTRUCTION ===
	subStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Muted))
	subtitle := subStyle.Render("S O U R C E   S E L E C T O R")
	instruction := subStyle.Render("Choose your contribution data source:")

	// === 3. SOURCE LIST OPTIONS ===
	var listLines []string
	for i, opt := range m.options {
		isSelected := (i == m.cursor)
		if isSelected {
			boxWidth := 46
			titlePart := fmt.Sprintf("› %s", opt.Title)
			titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cfg.Foreground))

			descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Muted))
			descPart := descStyle.Render(opt.Description)

			boxContent := fmt.Sprintf("%s\n  %s", titleStyle.Render(titlePart), descPart)

			boxStyle := lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(cfg.Accent)).
				Width(boxWidth).
				Padding(0, 1)

			renderedBox := boxStyle.Render(boxContent)
			listLines = append(listLines, strings.Split(renderedBox, "\n")...)
		} else {
			unselStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Foreground))
			descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Muted))
			listLines = append(listLines, fmt.Sprintf("    %s", unselStyle.Render(opt.Title)))
			listLines = append(listLines, fmt.Sprintf("    %s", descStyle.Render(opt.Description)))
			listLines = append(listLines, "")
		}
	}

	// === 4. BOTTOM HORIZONTAL DIVIDER ===
	ruleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Border))
	hDivider := ruleStyle.Render(strings.Repeat("─", totalWidth))

	// === 5. FOOTER ===
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cfg.Accent))
	txtStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Muted))

	footerLeft := fmt.Sprintf("%s %s   %s %s   %s %s   %s %s",
		keyStyle.Render("↑/↓"), txtStyle.Render("navigate"),
		keyStyle.Render("j/k"), txtStyle.Render("also work"),
		keyStyle.Render("Enter"), txtStyle.Render("select"),
		keyStyle.Render("Esc"), txtStyle.Render("cancel"),
	)

	margin := "  "
	var allLines []string
	allLines = append(allLines, "")
	for _, l := range wLines {
		allLines = append(allLines, margin+l)
	}
	allLines = append(allLines, "")
	allLines = append(allLines, margin+subtitle)
	allLines = append(allLines, margin+instruction)
	allLines = append(allLines, "")
	for _, l := range listLines {
		allLines = append(allLines, margin+l)
	}
	allLines = append(allLines, margin+hDivider)
	allLines = append(allLines, margin+footerLeft)
	allLines = append(allLines, "")

	return strings.Join(allLines, "\n")
}

// RunSourceSelector launches the interactive Bubble Tea source selector.
func RunSourceSelector(in io.Reader, out io.Writer) (string, bool, error) {
	model := NewSourceSelectorModel()
	opts := []tea.ProgramOption{
		tea.WithAltScreen(),
	}
	if in != nil {
		opts = append(opts, tea.WithInput(in))
	}
	if out != nil {
		opts = append(opts, tea.WithOutput(out))
	}

	p := tea.NewProgram(model, opts...)
	finalModel, err := p.Run()
	if err != nil {
		return "", true, err
	}

	res, ok := finalModel.(SourceSelectorModel)
	if !ok || res.Canceled || res.SelectedSource == "" {
		return "", true, nil
	}

	return res.SelectedSource, false, nil
}
