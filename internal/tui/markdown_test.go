package tui

import (
	"strings"
	"testing"
)

func TestRenderMarkdown_LightBackgroundUsesDarkText(t *testing.T) {
	const darkText = "\x1b[38;5;234m"

	got := renderMarkdownForBackground("Main task body", 80, false)
	if !strings.Contains(got, darkText) {
		t.Fatalf("light-background body color = %q, want ANSI color 234", got)
	}
}
