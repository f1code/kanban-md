package tui

import (
	"strings"
	"testing"
)

func TestRenderMarkdown_BodyUsesTerminalDefaultForeground(t *testing.T) {
	const (
		lightPaletteText = "\x1b[38;5;234m"
		darkPaletteText  = "\x1b[38;5;252m"
	)

	light := renderMarkdownForBackground("Main task body", 80, false)
	dark := renderMarkdownForBackground("Main task body", 80, true)

	for name, rendered := range map[string]string{
		"light": light,
		"dark":  dark,
	} {
		if strings.Contains(rendered, lightPaletteText) || strings.Contains(rendered, darkPaletteText) {
			t.Errorf("%s-background body uses a fixed foreground: %q", name, rendered)
		}
	}

	if light != dark {
		t.Errorf("plain body output depends on the startup background:\nlight: %q\ndark:  %q", light, dark)
	}
}
