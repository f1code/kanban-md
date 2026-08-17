package board

import (
	"testing"

	"github.com/antopolskiy/kanban-md/internal/config"
	"github.com/antopolskiy/kanban-md/internal/task"
)

func TestSummarizeChildrenDirectActiveChildrenInIDOrder(t *testing.T) {
	parentID := 10
	childID := 2
	tasks := []*task.Task{
		{ID: 5, Title: "Later child", Status: "todo", Parent: &parentID},
		{ID: 1, Title: "Unrelated", Status: "done"},
		{ID: 2, Title: "Earlier done child", Status: "done", Parent: &parentID},
		{ID: 3, Title: "Grandchild", Status: "backlog", Parent: &childID},
		{ID: 4, Title: "Archived child", Status: config.ArchivedStatus, Parent: &parentID},
	}

	summary := SummarizeChildren(tasks, parentID, config.NewDefault("test"), false)

	if summary.Done != 1 || summary.Total() != 2 {
		t.Fatalf("roll-up = %d/%d done, want 1/2", summary.Done, summary.Total())
	}
	if summary.Children[0].ID != 2 || summary.Children[1].ID != 5 {
		t.Fatalf("child IDs = [%d %d], want [2 5]", summary.Children[0].ID, summary.Children[1].ID)
	}
}

func TestSummarizeChildrenCanIncludeArchived(t *testing.T) {
	parentID := 10
	tasks := []*task.Task{
		{ID: 2, Title: "Done child", Status: "done", Parent: &parentID},
		{ID: 3, Title: "Archived child", Status: config.ArchivedStatus, Parent: &parentID},
	}

	summary := SummarizeChildren(tasks, parentID, config.NewDefault("test"), true)

	if summary.Done != 2 || summary.Total() != 2 {
		t.Fatalf("roll-up = %d/%d done, want 2/2", summary.Done, summary.Total())
	}
	if summary.Children[1].Status != config.ArchivedStatus {
		t.Fatalf("second child status = %q, want archived", summary.Children[1].Status)
	}
}

func TestSummarizeChildrenReturnsStableEmptySlice(t *testing.T) {
	selfID := 10
	tasks := []*task.Task{{ID: selfID, Title: "Self reference", Status: "todo", Parent: &selfID}}
	summary := SummarizeChildren(tasks, selfID, config.NewDefault("test"), false)
	if summary.Children == nil {
		t.Fatal("Children = nil, want a stable empty slice")
	}
	if summary.Done != 0 || summary.Total() != 0 {
		t.Fatalf("roll-up = %d/%d done, want 0/0", summary.Done, summary.Total())
	}
}

func TestFindParentResolvesArchivedParent(t *testing.T) {
	parentID := 4
	current := &task.Task{ID: 6, Parent: &parentID}
	tasks := []*task.Task{
		{ID: 1, Title: "Unrelated", Status: "todo"},
		{ID: 4, Title: "Archived parent", Status: config.ArchivedStatus},
	}

	parent := FindParent(tasks, current)
	if parent == nil {
		t.Fatal("FindParent() = nil, want archived parent")
	}
	if parent.ID != 4 || parent.Title != "Archived parent" || parent.Status != config.ArchivedStatus {
		t.Fatalf("FindParent() = %#v", parent)
	}
}

func TestFindParentMissingOrSelfReference(t *testing.T) {
	missingID := 99
	if parent := FindParent(nil, &task.Task{ID: 1, Parent: &missingID}); parent != nil {
		t.Fatalf("missing parent = %#v, want nil", parent)
	}
	selfID := 1
	if parent := FindParent([]*task.Task{{ID: 1}}, &task.Task{ID: 1, Parent: &selfID}); parent != nil {
		t.Fatalf("self parent = %#v, want nil", parent)
	}
}
