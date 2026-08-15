package board

import (
	"sort"

	"github.com/antopolskiy/kanban-md/internal/config"
	"github.com/antopolskiy/kanban-md/internal/task"
)

// ChildTask is the stable, read-only child representation used by task-detail views.
type ChildTask struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

// ChildSummary contains direct children and their terminal-status roll-up.
type ChildSummary struct {
	Children []ChildTask
	Done     int
}

// Total returns the number of direct children in the summary.
func (s ChildSummary) Total() int {
	return len(s.Children)
}

// SummarizeChildren returns direct children in ascending task-ID order.
// Archived children are omitted unless includeArchived is true. Self-references
// are ignored defensively; normal CLI mutations reject them before writing.
func SummarizeChildren(
	tasks []*task.Task,
	parentID int,
	cfg *config.Config,
	includeArchived bool,
) ChildSummary {
	summary := ChildSummary{Children: make([]ChildTask, 0)}
	for _, candidate := range tasks {
		if candidate.ID == parentID || candidate.Parent == nil || *candidate.Parent != parentID {
			continue
		}
		if !includeArchived && cfg.IsArchivedStatus(candidate.Status) {
			continue
		}

		summary.Children = append(summary.Children, ChildTask{
			ID:     candidate.ID,
			Title:  candidate.Title,
			Status: candidate.Status,
		})
		if cfg.IsTerminalStatus(candidate.Status) {
			summary.Done++
		}
	}

	sort.Slice(summary.Children, func(i, j int) bool {
		return summary.Children[i].ID < summary.Children[j].ID
	})
	return summary
}
