package render

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dexisback/gheppo/internal/stats"
)

// Model is the Bubble Tea presentation model managing interactive animation and layout.
type Model struct {
	summary       *stats.Summary
	layout        Layout
	theme         Theme
	wordmarkState WordmarkState
	progress      float64
	step          int
	totalSteps    int
	animating     bool
	quitting      bool
	width         int
	height        int
}

// TickMsg signals an animation frame advance.
type TickMsg time.Time

// NewModel initializes a Bubble Tea presentation model for the given contribution summary.
func NewModel(s *stats.Summary) Model {
	mode := DetectColorMode()
	themeMode := DetectTheme()
	theme := GetTheme(mode, themeMode)
	termWidth := GetTerminalWidth()

	totalWeeks := 0
	if s != nil && s.Grid != nil {
		totalWeeks = len(s.Grid)
	}

	layout := CalculateLayout(termWidth, totalWeeks)
	animating := ShouldAnimate()

	initialProgress := 0.0
	if !animating {
		initialProgress = 1.0
	}

	return Model{
		summary:       s,
		layout:        layout,
		theme:         theme,
		wordmarkState: WordmarkState{Progress: initialProgress},
		progress:      initialProgress,
		step:          0,
		totalSteps:    8,
		animating:     animating,
		quitting:      false,
		width:         termWidth,
	}
}

// Init begins the animation tick sequence.
func (m Model) Init() tea.Cmd {
	if !m.animating {
		return tea.Quit
	}
	return m.tickCmd()
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(18*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Update processes Bubble Tea messages (window resize, keypress, animation ticks).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		totalWeeks := 0
		if m.summary != nil && m.summary.Grid != nil {
			totalWeeks = len(m.summary.Grid)
		}
		m.layout = CalculateLayout(msg.Width, totalWeeks)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		}

	case TickMsg:
		m.step++
		m.progress = float64(m.step) / float64(m.totalSteps)
		if m.progress > 1.0 {
			m.progress = 1.0
		}
		m.wordmarkState.Progress = m.progress

		if m.step >= m.totalSteps {
			m.animating = false
			m.progress = 1.0
			m.wordmarkState.Progress = 1.0
			return m, tea.Quit
		}

		return m, m.tickCmd()
	}

	return m, nil
}

// View renders the presentation frame according to current model state.
func (m Model) View() string {
	if m.summary == nil {
		return ""
	}

	var frame string

	if m.animating {
		revealedWeeks := int(float64(m.layout.VisibleWeeks) * m.progress)
		if revealedWeeks < 1 {
			revealedWeeks = 1
		}
		stepSummary := createStepSummary(m.summary, m.layout.VisibleWeeks, revealedWeeks)
		frame = RenderWithThemeAndState(stepSummary, m.layout, m.theme, m.wordmarkState)
	} else {
		frame = RenderWithThemeAndState(m.summary, m.layout, m.theme, WordmarkState{Progress: 1.0})
	}

	// Terminate the frame with a newline: Bubble Tea's exit sequence erases the
	// line the cursor sits on, so without a trailing blank line the shell prompt
	// redraw would wipe the card's bottom row (the Saturday graph row).
	return frame + "\n"
}
