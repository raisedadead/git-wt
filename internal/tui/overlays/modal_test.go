package overlays

import (
	"strings"
	"testing"
)

func TestRenderModal(t *testing.T) {
	tests := []struct {
		name      string
		baseView  string
		modalView string
		width     int
		height    int
		check     func(t *testing.T, result string)
	}{
		{
			name:      "base content is dimmed",
			baseView:  "\033[31mRed text\033[0m\n\033[32mGreen text\033[0m\nPlain text",
			modalView: "X",
			width:     40,
			height:    5,
			check: func(t *testing.T, result string) {
				// Original ANSI color codes (e.g., \033[31m for red) should be stripped.
				// The result should NOT contain the original red/green escape codes.
				if strings.Contains(result, "\033[31m") {
					t.Error("expected original red ANSI code to be stripped from base")
				}
				if strings.Contains(result, "\033[32m") {
					t.Error("expected original green ANSI code to be stripped from base")
				}
			},
		},
		{
			name:      "modal content appears in output",
			baseView:  "line1\nline2\nline3\nline4\nline5",
			modalView: "+---------+\n| CONFIRM |\n+---------+",
			width:     40,
			height:    5,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "CONFIRM") {
					t.Error("expected modal content 'CONFIRM' to appear in output")
				}
				if !strings.Contains(result, "+---------+") {
					t.Error("expected modal border to appear in output")
				}
			},
		},
		{
			name:      "non-modal lines show base content",
			baseView:  "AAAA\nBBBB\nCCCC\nDDDD\nEEEE\nFFFF\nGGGG",
			modalView: "X",
			width:     40,
			height:    7,
			check: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				foundBase := false
				for _, line := range lines {
					stripped := strings.TrimSpace(line)
					if strings.Contains(stripped, "AAAA") ||
						strings.Contains(stripped, "GGGG") {
						foundBase = true
						break
					}
				}
				if !foundBase {
					t.Error("expected non-modal lines to contain base content (AAAA or GGGG)")
				}
			},
		},
		{
			name:      "empty base view",
			baseView:  "",
			modalView: "+---------+\n| CONFIRM |\n+---------+",
			width:     40,
			height:    5,
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "CONFIRM") {
					t.Error("expected modal to render on empty base")
				}
				lines := strings.Split(result, "\n")
				if len(lines) > 5 {
					t.Errorf("expected at most %d lines, got %d", 5, len(lines))
				}
			},
		},
		{
			name:      "modal taller than viewport",
			baseView:  "A\nB\nC",
			modalView: "1\n2\n3\n4\n5\n6\n7\n8\n9\n10",
			width:     40,
			height:    5,
			check: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				if len(lines) > 5 {
					t.Errorf("expected at most %d lines, got %d", 5, len(lines))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderModal(tt.baseView, tt.modalView, tt.width, tt.height)
			tt.check(t, result)
		})
	}
}
