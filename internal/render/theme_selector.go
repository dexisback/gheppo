package render

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dexisback/gheppo/internal/config"
)

// RunThemeSelector executes the interactive Bubble Tea theme selector session.
// Returns the chosen theme, whether the user canceled, and any error encountered.
func RunThemeSelector(in io.Reader, out io.Writer) (config.Theme, bool, error) {
	model := NewThemeSelectorModel()
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
		return config.Theme{}, true, err
	}

	res, ok := finalModel.(ThemeSelectorModel)
	if !ok || res.Canceled || res.AppliedTheme == "" {
		return config.Theme{}, true, nil
	}

	chosen := config.ResolveTheme(res.AppliedTheme)
	return chosen, false, nil
}

// ThemeSelectorModel manages the interactive theme selection state and rendering.
type ThemeSelectorModel struct {
	themes       []config.Theme
	cursor       int
	initialTheme string
	AppliedTheme string
	Canceled     bool
	quitting     bool
	width        int
	height       int
	mode         ColorMode
}

// NewThemeSelectorModel initializes a ThemeSelectorModel populated with registered themes.
func NewThemeSelectorModel() ThemeSelectorModel {
	themes := config.ListThemes()
	selectedName := config.GetSelectedThemeName()

	cursor := 0
	for i, t := range themes {
		if strings.EqualFold(t.Name, selectedName) {
			cursor = i
			break
		}
	}

	mode := DetectColorMode()
	termWidth := GetTerminalWidth()

	return ThemeSelectorModel{
		themes:       themes,
		cursor:       cursor,
		initialTheme: selectedName,
		width:        termWidth,
		height:       24,
		mode:         mode,
	}
}

// Init starts the Bubble Tea lifecycle.
func (m ThemeSelectorModel) Init() tea.Cmd {
	return nil
}

// Update handles keyboard navigation, selection, and window resizing.
func (m ThemeSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if len(m.themes) > 0 {
				m.cursor = (m.cursor - 1 + len(m.themes)) % len(m.themes)
			}
			return m, nil

		case "down", "j":
			if len(m.themes) > 0 {
				m.cursor = (m.cursor + 1) % len(m.themes)
			}
			return m, nil

		case "enter":
			if len(m.themes) > 0 {
				m.AppliedTheme = m.themes[m.cursor].Name
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

// View constructs the complete interactive selector UI matching the design specification.
func (m ThemeSelectorModel) View() string {
	if m.quitting {
		return ""
	}

	if len(m.themes) == 0 {
		return "No themes available.\n"
	}

	activeCfg := m.themes[m.cursor]
	activeTheme := ResolveTheme(activeCfg, m.mode)

	// Layout dimensions
	leftWidth := 46
	rightWidth := 34
	totalWidth := leftWidth + 3 + rightWidth // left + divider (3) + right = 83

	// === 1. BUILD LEFT COLUMN ===
	var leftLines []string

	// Wordmark
	wLines, _ := RenderWordmarkWithStateAndWidth(activeTheme, WordmarkState{Progress: 1.0}, leftWidth)
	leftLines = append(leftLines, wLines...)

	// Spacing
	leftLines = append(leftLines, "")

	// Subtitle & Instruction
	subStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(activeCfg.Muted))
	leftLines = append(leftLines, subStyle.Render("T H E M E   S E L E C T O R"))
	leftLines = append(leftLines, subStyle.Render("Choose a theme for your contribution graph."))
	leftLines = append(leftLines, "")

	// Theme list options
	for i, t := range m.themes {
		isSelected := (i == m.cursor)
		if isSelected {
			// Build selected box with swatches
			boxWidth := 44
			titlePart := fmt.Sprintf("› %s", t.Title())
			titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(activeCfg.Foreground))

			// Build 5 palette swatches
			var swatches strings.Builder
			for lvl := 0; lvl < 5; lvl++ {
				colorHex := t.GraphLevels[lvl]
				swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(colorHex)).Render("■")
				if lvl > 0 {
					swatches.WriteString(" ")
				}
				swatches.WriteString(swatch)
			}

			innerSpace := boxWidth - 4 - lipgloss.Width(titlePart) - 9
			if innerSpace < 1 {
				innerSpace = 1
			}

			boxContent := fmt.Sprintf("%s%s%s",
				titleStyle.Render(titlePart),
				strings.Repeat(" ", innerSpace),
				swatches.String(),
			)

			boxStyle := lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(activeCfg.Accent)).
				Width(boxWidth).
				Padding(0, 1)

			renderedBox := boxStyle.Render(boxContent)
			leftLines = append(leftLines, strings.Split(renderedBox, "\n")...)
		} else {
			unselStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(activeCfg.Foreground))
			leftLines = append(leftLines, fmt.Sprintf("    %s", unselStyle.Render(t.Title())))
			leftLines = append(leftLines, "")
		}
	}

	// === 2. BUILD RIGHT COLUMN ===
	var rightLines []string

	// Title
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(activeCfg.Accent))
	rightLines = append(rightLines, nameStyle.Render(activeCfg.Title()))

	// Description
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(activeCfg.Muted)).Width(rightWidth)
	desc := activeCfg.Description
	if desc == "" {
		desc = "Contribution graph palette."
	}
	descRendered := descStyle.Render(desc)
	rightLines = append(rightLines, strings.Split(descRendered, "\n")...)

	// Separator rule
	ruleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(activeCfg.Border))
	rightLines = append(rightLines, ruleStyle.Render(strings.Repeat("─", rightWidth-4)))
	rightLines = append(rightLines, "")

	// 5 Levels Breakdown
	for lvl := 0; lvl < 5; lvl++ {
		colorHex := activeCfg.GraphLevels[lvl]
		swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(colorHex)).Render("■")
		label := fmt.Sprintf("Level %d", lvl)
		labelStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(activeCfg.Muted)).Render(label)
		hexStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(activeCfg.Muted)).Render(colorHex)

		row := fmt.Sprintf("%s  %-10s %s", swatch, labelStyled, hexStyled)
		rightLines = append(rightLines, row)
	}

	// Align rows between left and right columns
	// In the design, the right column starts at row 6 (aligned with the theme list)
	rightOffset := 6
	paddedRightLines := make([]string, rightOffset)
	for i := 0; i < rightOffset; i++ {
		paddedRightLines[i] = ""
	}
	paddedRightLines = append(paddedRightLines, rightLines...)

	maxRows := len(leftLines)
	if len(paddedRightLines) > maxRows {
		maxRows = len(paddedRightLines)
	}

	// Pad lines to consistent heights
	for len(leftLines) < maxRows {
		leftLines = append(leftLines, "")
	}
	for len(paddedRightLines) < maxRows {
		paddedRightLines = append(paddedRightLines, "")
	}

	// === 3. BUILD VERTICAL DIVIDER & JOIN BODY ===
	divCell := activeTheme.Border + "│" + activeTheme.Reset
	gapDivider := "  " + divCell + "  "
	emptyDivider := "     "

	// Format left column strings to exact width
	for i := range leftLines {
		visLen := lipgloss.Width(leftLines[i])
		if visLen < leftWidth {
			leftLines[i] += strings.Repeat(" ", leftWidth-visLen)
		}
	}

	dividerStartRow := 5
	var bodyLines []string
	for i := 0; i < maxRows; i++ {
		div := emptyDivider
		if i >= dividerStartRow {
			div = gapDivider
		}
		bodyLines = append(bodyLines, leftLines[i]+div+paddedRightLines[i])
	}
	body := strings.Join(bodyLines, "\n")

	// === 4. BOTTOM HORIZONTAL DIVIDER ===
	hDivider := ruleStyle.Render(strings.Repeat("─", totalWidth))

	// === 5. FOOTER (KEYBINDINGS & TAGLINE) ===
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(activeCfg.Accent))
	txtStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(activeCfg.Muted))

	footerLeft := fmt.Sprintf("%s %s   %s %s   %s %s   %s %s",
		keyStyle.Render("↑/↓"), txtStyle.Render("navigate"),
		keyStyle.Render("j/k"), txtStyle.Render("also work"),
		keyStyle.Render("Enter"), txtStyle.Render("apply"),
		keyStyle.Render("Esc"), txtStyle.Render("cancel"),
	)

	tagline := txtStyle.Render("Make your graph yours.")
	footerLeftWidth := lipgloss.Width(footerLeft)
	taglineWidth := lipgloss.Width(tagline)

	footerPad := totalWidth - footerLeftWidth - taglineWidth
	if footerPad < 2 {
		footerPad = 2
	}
	footer := footerLeft + strings.Repeat(" ", footerPad) + tagline

	// Apply left indentation margin
	margin := "  "
	finalLines := []string{
		"",
		margin + strings.ReplaceAll(body, "\n", "\n"+margin),
		margin + hDivider,
		margin + footer,
		"",
	}

	return strings.Join(finalLines, "\n")
}
