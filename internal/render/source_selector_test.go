package render

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dexisback/gheppo/internal/config"
)

func TestSourceSelectorModelNavigation(t *testing.T) {
	m := NewSourceSelectorModel()
	if len(m.options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(m.options))
	}

	initialCursor := m.cursor

	// Move down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(SourceSelectorModel)
	if m.cursor != (initialCursor+1)%2 {
		t.Errorf("cursor after down = %d, want %d", m.cursor, (initialCursor+1)%2)
	}

	// Move down with 'j'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(SourceSelectorModel)
	if m.cursor != initialCursor {
		t.Errorf("cursor after j = %d, want %d", m.cursor, initialCursor)
	}

	// Move up with 'k'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(SourceSelectorModel)
	if m.cursor != (initialCursor+1)%2 {
		t.Errorf("cursor after k = %d, want %d", m.cursor, (initialCursor+1)%2)
	}

	// Move up with KeyUp
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(SourceSelectorModel)
	if m.cursor != initialCursor {
		t.Errorf("cursor after up = %d, want %d", m.cursor, initialCursor)
	}
}

func TestSourceSelectorModelApply(t *testing.T) {
	m := NewSourceSelectorModel()
	m.cursor = 1 // LeetCode

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(SourceSelectorModel)

	if !m.quitting {
		t.Error("expected quitting = true on Enter")
	}
	if m.Canceled {
		t.Error("expected Canceled = false on Enter")
	}
	if m.SelectedSource != config.SourceLeetCode {
		t.Errorf("SelectedSource = %q, want %q", m.SelectedSource, config.SourceLeetCode)
	}
	if cmd == nil {
		t.Error("expected tea.Quit command on Enter")
	}
}

func TestSourceSelectorModelCancel(t *testing.T) {
	for _, key := range []string{"esc", "q", "ctrl+c"} {
		m := NewSourceSelectorModel()
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
		m = updated.(SourceSelectorModel)

		if !m.quitting {
			t.Errorf("key %q: expected quitting = true", key)
		}
		if !m.Canceled {
			t.Errorf("key %q: expected Canceled = true", key)
		}
		if m.SelectedSource != "" {
			t.Errorf("key %q: expected empty SelectedSource, got %q", key, m.SelectedSource)
		}
		if cmd == nil {
			t.Errorf("key %q: expected tea.Quit command", key)
		}
	}
}

func TestSourceSelectorModelView(t *testing.T) {
	m := NewSourceSelectorModel()

	for i := range m.options {
		m.cursor = i
		view := m.View()

		if !strings.Contains(view, "S O U R C E   S E L E C T O R") {
			t.Errorf("option %d view missing header: %q", i, view)
		}
		if !strings.Contains(view, "GitHub") || !strings.Contains(view, "LeetCode") {
			t.Errorf("option %d view missing sources: %q", i, view)
		}
		if !strings.Contains(view, "navigate") || !strings.Contains(view, "select") || !strings.Contains(view, "cancel") {
			t.Errorf("option %d view missing footer keys: %q", i, view)
		}
	}
}

func TestRunSourceSelectorSimulation(t *testing.T) {
	input := strings.NewReader("\r")
	out := new(strings.Builder)

	selected, canceled, err := RunSourceSelector(input, out)
	if err != nil {
		t.Fatalf("RunSourceSelector failed: %v", err)
	}
	if canceled {
		t.Error("expected canceled = false on Enter")
	}
	if selected == "" {
		t.Error("expected non-empty selected source")
	}
}
