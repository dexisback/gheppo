package render

import (
	"testing"
)

func TestCalculateLayout(t *testing.T) {
	tests := []struct {
		name           string
		termWidth      int
		totalWeeks     int
		wantCardW      int
		wantVisWeeks   int
		wantDividerW   int
	}{
		{
			name:         "narrow terminal clamped to min",
			termWidth:    30,
			totalWeeks:   52,
			wantCardW:    MinCardWidth,
			wantVisWeeks: 10,
			wantDividerW: 3,
		},
		{
			name:         "standard 80 col terminal",
			termWidth:    80,
			totalWeeks:   52,
			wantCardW:    79,
			wantVisWeeks: 20,
			wantDividerW: 5,
		},
		{
			name:         "wide terminal scaled",
			termWidth:    120,
			totalWeeks:   52,
			wantCardW:    119,
			wantVisWeeks: 35,
			wantDividerW: 5,
		},
		{
			name:         "limited available data weeks",
			termWidth:    100,
			totalWeeks:   10,
			wantCardW:    99,
			wantVisWeeks: 10,
			wantDividerW: 5,
		},
		{
			name:         "zero terminal width defaults to 80",
			termWidth:    0,
			totalWeeks:   52,
			wantCardW:    79,
			wantVisWeeks: 20,
			wantDividerW: 5,
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
			if l.DividerWidth != tt.wantDividerW {
				t.Errorf("DividerWidth = %d, want %d", l.DividerWidth, tt.wantDividerW)
			}
		})
	}
}

func TestLayoutLeftMarginCentering(t *testing.T) {
	l80 := CalculateLayout(80, 52)
	if l80.LeftMargin != 4 {
		t.Errorf("LeftMargin = %d, want 4", l80.LeftMargin)
	}
}
