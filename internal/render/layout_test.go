package render

import (
	"testing"
)

func TestCalculateLayout(t *testing.T) {
	tests := []struct {
		name         string
		termWidth    int
		totalWeeks   int
		wantCardW    int
		wantVisWeeks int
	}{
		{
			name:         "narrow terminal clamped to min",
			termWidth:    30,
			totalWeeks:   52,
			wantCardW:    MinCardWidth,
			wantVisWeeks: 9, // inner: 42, divider: 3 -> content: 39 -> stats min: 20 -> graph: 19 -> 19/2 = 9
		},
		{
			name:         "standard 80 col terminal",
			termWidth:    80,
			totalWeeks:   52,
			wantCardW:    76,
			wantVisWeeks: 24,
		},
		{
			name:         "wide terminal clamped to max",
			termWidth:    200,
			totalWeeks:   52,
			wantCardW:    MaxCardWidth,
			wantVisWeeks: 46,
		},
		{
			name:         "limited available data weeks",
			termWidth:    100,
			totalWeeks:   10,
			wantCardW:    96,
			wantVisWeeks: 10,
		},
		{
			name:         "zero terminal width defaults to 80",
			termWidth:    0,
			totalWeeks:   52,
			wantCardW:    76,
			wantVisWeeks: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := CalculateLayout(tt.termWidth, tt.totalWeeks)
			if l.CardWidth != tt.wantCardW {
				t.Errorf("CardWidth = %d, want %d", l.CardWidth, tt.wantCardW)
			}
			if l.VisibleWeeks != tt.wantVisWeeks {
				t.Errorf("VisibleWeeks = %d, want %d", l.VisibleWeeks, tt.wantVisWeeks)
			}
			if l.GraphWidth != l.VisibleWeeks*2 {
				t.Errorf("GraphWidth = %d, want %d", l.GraphWidth, l.VisibleWeeks*2)
			}
			if l.GraphHeight != 7 {
				t.Errorf("GraphHeight = %d, want 7", l.GraphHeight)
			}
			if l.DividerWidth != 3 {
				t.Errorf("DividerWidth = %d, want 3", l.DividerWidth)
			}
		})
	}
}

func TestLayoutLeftMarginCentering(t *testing.T) {
	// 120 col terminal with 116 card width -> margin = (120-116)/2 = 2
	l := CalculateLayout(120, 52)
	expectedMargin := (120 - l.CardWidth) / 2
	if l.LeftMargin != expectedMargin {
		t.Errorf("LeftMargin = %d, want %d", l.LeftMargin, expectedMargin)
	}

	// 80 col terminal with 76 card width -> margin = (80-76)/2 = 2
	l80 := CalculateLayout(80, 52)
	if l80.LeftMargin != 2 {
		t.Errorf("LeftMargin = %d, want 2", l80.LeftMargin)
	}
}
