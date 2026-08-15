package tui

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeTestExecutable(t *testing.T, name string) string {
	t.Helper()

	contents := "#!/bin/sh\nexit 0\n"
	if runtime.GOOS == "windows" {
		name += ".bat"
		contents = "@exit /b 0\r\n"
		t.Setenv("PATHEXT", ".BAT;.CMD;.EXE")
	}

	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("writing test executable: %v", err)
	}
	if err := os.Chmod(path, 0o700); err != nil { //nolint:gosec // LookPath requires an executable test file
		t.Fatalf("making test executable: %v", err)
	}
	return path
}

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
	t.Setenv("PATH", t.TempDir())

	if _, err := externalEditorCommand("001-task.md"); err == nil {
		t.Fatal("externalEditorCommand() error = nil, want missing editor error")
	} else if !strings.Contains(err.Error(), "vi") {
		t.Fatalf("externalEditorCommand() error = %q, want vi fallback guidance", err)
	}
}

func TestExternalEditorCommandFallsBackToVi(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	viPath := writeTestExecutable(t, "vi")
	t.Setenv("PATH", filepath.Dir(viPath))

	taskPath := filepath.Join(t.TempDir(), "001-task.md")
	cmd, err := externalEditorCommand(taskPath)
	if err != nil {
		t.Fatalf("externalEditorCommand() error = %v", err)
	}
	if got := cmd.Args; len(got) != 2 || got[0] != viPath || got[1] != taskPath {
		t.Fatalf("command args = %q, want [%s %s]", got, viPath, taskPath)
	}
}
