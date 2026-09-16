package render

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestThemeSelectorModelNavigation(t *testing.T) {
	m := NewThemeSelectorModel()
	if len(m.themes) == 0 {
		t.Fatal("expected registered themes in selector model")
	}

	initialCursor := m.cursor

	// Move down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(ThemeSelectorModel)
	expectedDown := (initialCursor + 1) % len(m.themes)
	if m.cursor != expectedDown {
		t.Errorf("cursor after down = %d, want %d", m.cursor, expectedDown)
	}

	// Move down with 'j'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(ThemeSelectorModel)
	expectedJ := (expectedDown + 1) % len(m.themes)
	if m.cursor != expectedJ {
		t.Errorf("cursor after j = %d, want %d", m.cursor, expectedJ)
	}

	// Move up with 'k'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(ThemeSelectorModel)
	if m.cursor != expectedDown {
		t.Errorf("cursor after k = %d, want %d", m.cursor, expectedDown)
	}

	// Move up with KeyUp
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(ThemeSelectorModel)
	if m.cursor != initialCursor {
		t.Errorf("cursor after up = %d, want %d", m.cursor, initialCursor)
	}
}

func TestThemeSelectorModelApply(t *testing.T) {
	m := NewThemeSelectorModel()
	m.cursor = 3 // Nord (assuming github, mono, catppuccin, nord, gruvbox)
	expectedTheme := m.themes[3].Name

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(ThemeSelectorModel)

	if !m.quitting {
		t.Error("expected quitting = true on Enter")
	}
	if m.Canceled {
		t.Error("expected Canceled = false on Enter")
	}
	if m.AppliedTheme != expectedTheme {
		t.Errorf("AppliedTheme = %q, want %q", m.AppliedTheme, expectedTheme)
	}
	if cmd == nil {
		t.Error("expected tea.Quit command on Enter")
	}
}

func TestThemeSelectorModelCancel(t *testing.T) {
	for _, key := range []string{"esc", "q", "ctrl+c"} {
		m := NewThemeSelectorModel()
		var msg tea.KeyMsg
		switch key {
		case "esc":
			msg = tea.KeyMsg{Type: tea.KeyEsc}
		case "q":
			msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		case "ctrl+c":
			msg = tea.KeyMsg{Type: tea.KeyCtrlC}
		}

		updated, cmd := m.Update(msg)
		m = updated.(ThemeSelectorModel)

		if !m.quitting {
			t.Errorf("key %q: expected quitting = true", key)
		}
		if !m.Canceled {
			t.Errorf("key %q: expected Canceled = true", key)
		}
		if m.AppliedTheme != "" {
			t.Errorf("key %q: expected empty AppliedTheme, got %q", key, m.AppliedTheme)
		}
		if cmd == nil {
			t.Errorf("key %q: expected tea.Quit command", key)
		}
	}
}

func TestThemeSelectorModelView(t *testing.T) {
	m := NewThemeSelectorModel()

	for i, themeCfg := range m.themes {
		m.cursor = i
		view := m.View()

		if !strings.Contains(view, "THEME   SELECTOR") && !strings.Contains(view, "THEME SELECTOR") && !strings.Contains(view, "T H E M E") {
			t.Errorf("theme %s view missing header: %q", themeCfg.Name, view)
		}
		if !strings.Contains(view, "Choose a theme for your contribution graph.") {
			t.Errorf("theme %s view missing instruction: %q", themeCfg.Name, view)
		}
		if !strings.Contains(view, themeCfg.Title()) {
			t.Errorf("theme %s view missing title: %q", themeCfg.Name, view)
		}
		if !strings.Contains(view, "Level 0") || !strings.Contains(view, "Level 4") {
			t.Errorf("theme %s view missing level labels: %q", themeCfg.Name, view)
		}
		if !strings.Contains(view, themeCfg.GraphLevels[0]) || !strings.Contains(view, themeCfg.GraphLevels[4]) {
			t.Errorf("theme %s view missing level hex codes: %q", themeCfg.Name, view)
		}
		if !strings.Contains(view, "navigate") || !strings.Contains(view, "apply") || !strings.Contains(view, "cancel") {
			t.Errorf("theme %s view missing footer navigation keys: %q", themeCfg.Name, view)
		}
		if !strings.Contains(view, "Make your graph yours.") {
			t.Errorf("theme %s view missing tagline: %q", themeCfg.Name, view)
		}
	}
}

func TestThemeSelectorModelWindowResize(t *testing.T) {
	m := NewThemeSelectorModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(ThemeSelectorModel)

	if m.width != 120 {
		t.Errorf("m.width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("m.height = %d, want 40", m.height)
	}
}

func TestRunThemeSelectorSimulation(t *testing.T) {
	// Simulate user pressing Enter immediately
	input := strings.NewReader("\r")
	out := new(strings.Builder)

	chosen, canceled, err := RunThemeSelector(input, out)
	if err != nil {
		t.Fatalf("RunThemeSelector failed: %v", err)
	}
	if canceled {
		t.Error("expected canceled = false on Enter")
	}
	if chosen.Name == "" {
		t.Error("expected non-empty chosen theme name")
	}
}
