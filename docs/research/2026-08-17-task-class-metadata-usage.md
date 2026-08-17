# Task `class` metadata usage

Date: 2026-08-17

## Question

What does the `class` field in task frontmatter represent, and does the application use it?

## Conclusion

`class` is the task's **class of service**: a workflow category that controls how urgently or specially a task should be handled. It is active application data, not inert descriptive metadata.

The default configuration defines the classes in this order:

1. `expedite`
2. `fixed-date`
3. `standard`
4. `intangible`

New tasks default to `standard`. The configured order is also the pick order. `expedite` has a default board-wide WIP limit of one and bypasses column WIP limits; `fixed-date` tasks use due dates as an additional ordering rule; `intangible` work is picked after the other classes.

## Where it is represented

- `internal/task/task.go`: `Task.Class` is serialized as the optional YAML/JSON `class` field.
- `internal/config/config.go`: board configuration contains an ordered `classes` list, per-class `wip_limit` and `bypass_column_wip` rules, and `defaults.class`.
- `internal/config/defaults.go`: new boards default to the four classes above and default new tasks to `standard`.
- `internal/board/mutate.go`: task creation applies `defaults.class`; explicit class values are validated against the configured class names.
- `cmd/create.go` and `cmd/edit.go`: `--class` creates or updates the value.

## Runtime behavior

### Automatic picking

`internal/board/pick.go` sorts eligible tasks by the order of `config.classes` before comparing task priority. With the default configuration, this means `expedite`, then `fixed-date`, then `standard`, then `intangible`. When both candidates are `fixed-date`, their due dates are compared before the normal priority fallback.

An absent or unknown class is treated as `standard` for pick ordering.

### WIP enforcement

`internal/board/mutate.go` uses a task's class during create and move operations:

- A positive per-class `wip_limit` is enforced board-wide.
- `bypass_column_wip: true` skips the target status column's WIP limit.
- Classes without either option use ordinary column WIP enforcement.

The default `expedite` definition has `wip_limit: 1` and `bypass_column_wip: true`.

### Querying and presentation

The field is also used by:

- `list --class` filtering (`internal/board/filter.go`)
- `board --group-by class` and `list --group-by class` grouping (`internal/board/group.go`)
- board class counts (`internal/board/board.go`)
- task detail and TUI detail rendering (`internal/output/table.go`, `internal/tui/board.go`)
- JSON serialization through the task schema

## Current kanban-md board

The canonical board config uses the default four classes and `defaults.class: standard`. At inspection time, `kanban-md board --table` reported:

| Class | Active/non-archived tasks |
|---|---:|
| expedite | 0 |
| fixed-date | 0 |
| standard | 239 |
| intangible | 0 |

Of the 249 task files including archived tasks, 195 explicitly contain `class: standard` and 54 older files omit the field. Therefore the field is implemented and behaviorally significant, but it does not currently distinguish work on this project's board: every task is effectively standard.

## Inconsistencies found

### Classless filtering

Pick ordering, grouping, and board summaries normalize a missing class to `standard`, but `list --class standard` performs an exact string comparison. It returned 185 non-archived tasks instead of the 239 shown in the standard board count because 54 active legacy tasks omit the field. This is tracked as task #255.

### Class WIP includes terminal tasks

`countByClass` counts tasks in every status. A temporary-board probe created an `expedite` task, moved it to `done`, then attempted to create a second `expedite` task. The second create failed with `expedite WIP limit reached (1/1 board-wide)`. Terminal work therefore consumes class WIP capacity. This is tracked as task #256.

