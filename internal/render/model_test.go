package render

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dexisback/gheppo/internal/stats"
)

func TestModelInitialization(t *testing.T) {
	summary := &stats.Summary{
		Login: "modeluser",
		Total: 50,
	}

	m := NewModel(summary)
	if m.summary.Login != "modeluser" {
		t.Errorf("expected login modeluser, got %s", m.summary.Login)
	}

	if m.totalSteps != 8 {
		t.Errorf("expected totalSteps 8, got %d", m.totalSteps)
	}

	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view")
	}

	if !strings.Contains(view, "@modeluser") {
		t.Errorf("view missing @modeluser: %s", view)
	}
}

func TestModelWindowResize(t *testing.T) {
	summary := &stats.Summary{
		Login: "resizeuser",
		Total: 100,
	}

	m := NewModel(summary)
	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if cmd != nil {
		t.Errorf("expected nil cmd on window resize, got %v", cmd)
	}

	updatedModel, ok := updated.(Model)
	if !ok {
		t.Fatalf("expected Model type, got %T", updated)
	}

	if updatedModel.width != 100 {
		t.Errorf("expected width 100, got %d", updatedModel.width)
	}
	if updatedModel.height != 40 {
		t.Errorf("expected height 40, got %d", updatedModel.height)
	}
}

func TestModelKeyQuit(t *testing.T) {
	summary := &stats.Summary{Login: "quituser"}
	m := NewModel(summary)

	keys := []string{"q", "esc", "ctrl+c"}
	for _, key := range keys {
		var keyMsg tea.KeyMsg
		switch key {
		case "ctrl+c":
			keyMsg = tea.KeyMsg{Type: tea.KeyCtrlC}
		case "esc":
			keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
		default:
			keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}

		updated, cmd := m.Update(keyMsg)
		if cmd == nil {
			t.Errorf("expected tea.Quit command for key %s", key)
		}
		updatedModel, ok := updated.(Model)
		if !ok || !updatedModel.quitting {
			t.Errorf("expected quitting true for key %s", key)
		}
	}
}

func TestModelAnimationTicks(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	summary := &stats.Summary{
		Login: "tickuser",
		Total: 200,
	}

	m := NewModel(summary)
	m.animating = true

	// Simulate 8 ticks
	var current tea.Model = m
	var cmd tea.Cmd
	for step := 1; step <= 8; step++ {
		current, cmd = current.Update(TickMsg(time.Now()))
		mod := current.(Model)
		if mod.step != step {
			t.Errorf("step = %d, want %d", mod.step, step)
		}
		if step < 8 && cmd == nil {
			t.Errorf("expected tick command on step %d", step)
		}
	}

	finalModel := current.(Model)
	if finalModel.animating {
		t.Error("expected animating false after 8 steps")
	}
	if finalModel.progress != 1.0 {
		t.Errorf("expected progress 1.0, got %f", finalModel.progress)
	}
}
