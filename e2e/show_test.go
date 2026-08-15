package e2e_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Show tests
// ---------------------------------------------------------------------------

func TestShow(t *testing.T) {
	kanbanDir := initBoard(t)
	mustCreateTask(t, kanbanDir, "Show me", "--body", "Detailed description",
		"--assignee", assigneeAlice, "--tags", "test")

	var task taskJSON
	r := runKanbanJSON(t, kanbanDir, &task, "show", "1")

	if r.exitCode != 0 {
		t.Fatalf("show failed: %s", r.stderr)
	}
	if task.ID != 1 {
		t.Errorf("ID = %d, want 1", task.ID)
	}
	if task.Title != "Show me" {
		t.Errorf("Title = %q, want %q", task.Title, "Show me")
	}
	if !strings.Contains(task.Body, "Detailed description") {
		t.Errorf("Body = %q, want to contain %q", task.Body, "Detailed description")
	}
	if task.Assignee != assigneeAlice {
		t.Errorf("Assignee = %q, want %q", task.Assignee, assigneeAlice)
	}
}

func TestShowNotFound(t *testing.T) {
	kanbanDir := initBoard(t)

	errResp := runKanbanJSONError(t, kanbanDir, "show", "999")
	if errResp.Code != "TASK_NOT_FOUND" {
		t.Errorf("code = %q, want TASK_NOT_FOUND", errResp.Code)
	}
}

func TestShowInvalidID(t *testing.T) {
	kanbanDir := initBoard(t)

	errResp := runKanbanJSONError(t, kanbanDir, "show", "abc")
	if errResp.Code != "INVALID_TASK_ID" {
		t.Errorf("code = %q, want INVALID_TASK_ID", errResp.Code)
	}
}

// ---------------------------------------------------------------------------
// Edit tests
// ---------------------------------------------------------------------------

func TestShowDisplaysLeadCycleTime(t *testing.T) {
	kanbanDir := initBoard(t)
	mustCreateTask(t, kanbanDir, "Task A")
	runKanban(t, kanbanDir, "--json", "move", "1", statusTodo)
	runKanban(t, kanbanDir, "--json", "move", "1", "done")

	r := runKanban(t, kanbanDir, "--table", "show", "1")
	if r.exitCode != 0 {
		t.Fatalf("show failed: %s", r.stderr)
	}
	if !strings.Contains(r.stdout, "Lead time") {
		t.Errorf("show output missing 'Lead time', got: %s", r.stdout)
	}
	if !strings.Contains(r.stdout, "Cycle time") {
		t.Errorf("show output missing 'Cycle time', got: %s", r.stdout)
	}
}

// ---------------------------------------------------------------------------
// Dependency tests
// ---------------------------------------------------------------------------

func TestShowCompactOutput(t *testing.T) {
	kanbanDir := initBoard(t)
	mustCreateTask(t, kanbanDir, "Compact show test")

	r := runKanban(t, kanbanDir, "show", "1", "--compact")
	if r.exitCode != 0 {
		t.Fatalf("show --compact failed (exit %d): %s", r.exitCode, r.stderr)
	}
	if !strings.Contains(r.stdout, "Compact show test") {
		t.Error("compact show output should contain task title")
	}
}

func seedShowChildrenBoard(t *testing.T) string {
	t.Helper()
	kanbanDir := initBoard(t)
	mustCreateTask(t, kanbanDir, "Parent epic", "--priority", "critical")
	mustCreateTask(t, kanbanDir, "Backlog child", "--parent", "1")
	mustCreateTask(t, kanbanDir, "Done child", "--parent", "1", "--status", "done")
	mustCreateTask(t, kanbanDir, "Archived child", "--parent", "1")
	archive := runKanban(t, kanbanDir, "archive", "4")
	if archive.exitCode != 0 {
		t.Fatalf("archiving child failed: %s", archive.stderr)
	}
	return kanbanDir
}

func TestShowChildrenTableAndCompact(t *testing.T) {
	kanbanDir := seedShowChildrenBoard(t)

	table := runKanban(t, kanbanDir, "show", "1")
	if table.exitCode != 0 {
		t.Fatalf("show failed: %s", table.stderr)
	}
	for _, want := range []string{"Children (1/2 done)", "#2", "Backlog child", "#3", "Done child"} {
		if !strings.Contains(table.stdout, want) {
			t.Errorf("table output missing %q:\n%s", want, table.stdout)
		}
	}
	if strings.Contains(table.stdout, "Archived child") {
		t.Errorf("default show should hide archived children:\n%s", table.stdout)
	}
	childTable := runKanban(t, kanbanDir, "show", "2")
	if !strings.Contains(childTable.stdout, "Parent:") || !strings.Contains(childTable.stdout, "#1") {
		t.Errorf("child detail should show its parent:\n%s", childTable.stdout)
	}

	compact := runKanban(t, kanbanDir, "show", "1", "--compact")
	if compact.exitCode != 0 {
		t.Fatalf("show --compact failed: %s", compact.stderr)
	}
	if !strings.Contains(compact.stdout, "children:1/2 done") {
		t.Errorf("compact output missing roll-up:\n%s", compact.stdout)
	}
	if strings.Contains(compact.stdout, "Archived child") {
		t.Errorf("compact output should hide archived children:\n%s", compact.stdout)
	}
}

func TestShowChildrenJSON(t *testing.T) {
	kanbanDir := seedShowChildrenBoard(t)

	var detail struct {
		taskJSON
		Children []struct {
			ID     int    `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"children"`
	}
	jsonResult := runKanbanJSON(t, kanbanDir, &detail, "show", "1")
	if jsonResult.exitCode != 0 {
		t.Fatalf("show --json failed: %s", jsonResult.stderr)
	}
	if len(detail.Children) != 2 || detail.Children[0].ID != 2 || detail.Children[1].ID != 3 {
		t.Fatalf("JSON children = %#v, want IDs 2 and 3", detail.Children)
	}
}

func TestShowChildrenArchived(t *testing.T) {
	kanbanDir := seedShowChildrenBoard(t)

	withArchived := runKanban(t, kanbanDir, "show", "1", "--archived")
	if withArchived.exitCode != 0 {
		t.Fatalf("show --archived failed: %s", withArchived.stderr)
	}
	if !strings.Contains(withArchived.stdout, "Children (2/3 done)") ||
		!strings.Contains(withArchived.stdout, "Archived child") {
		t.Errorf("show --archived should include archived child:\n%s", withArchived.stdout)
	}
}

func TestShowJSONAlwaysIncludesChildrenArray(t *testing.T) {
	kanbanDir := initBoard(t)
	mustCreateTask(t, kanbanDir, "Leaf task")

	r := runKanban(t, kanbanDir, "--json", "show", "1")
	if r.exitCode != 0 {
		t.Fatalf("show --json failed: %s", r.stderr)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(r.stdout), &raw); err != nil {
		t.Fatalf("unmarshal show JSON: %v", err)
	}
	if got := strings.TrimSpace(string(raw["children"])); got != "[]" {
		t.Fatalf("children JSON = %s, want []", got)
	}
}
