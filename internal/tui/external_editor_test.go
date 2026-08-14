package tui

import (
	"path/filepath"
	"testing"
)

func TestExternalEditorCommandPrefersVisual(t *testing.T) {
	t.Setenv("VISUAL", "  hx  ")
	t.Setenv("EDITOR", "vim")

	taskPath := filepath.Join(t.TempDir(), "001-task.md")
	cmd, err := externalEditorCommand(taskPath)
	if err != nil {
		t.Fatalf("externalEditorCommand() error = %v", err)
	}
	if got := cmd.Args; len(got) != 2 || got[0] != "hx" || got[1] != taskPath {
		t.Fatalf("command args = %q, want [hx %s]", got, taskPath)
	}
}

func TestExternalEditorCommandFallsBackToEditor(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "vim")

	taskPath := filepath.Join(t.TempDir(), "001-task.md")
	cmd, err := externalEditorCommand(taskPath)
	if err != nil {
		t.Fatalf("externalEditorCommand() error = %v", err)
	}
	if got := cmd.Args; len(got) != 2 || got[0] != "vim" || got[1] != taskPath {
		t.Fatalf("command args = %q, want [vim %s]", got, taskPath)
	}
}

func TestExternalEditorCommandRequiresEditor(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	if _, err := externalEditorCommand("001-task.md"); err == nil {
		t.Fatal("externalEditorCommand() error = nil, want missing editor error")
	}
}
