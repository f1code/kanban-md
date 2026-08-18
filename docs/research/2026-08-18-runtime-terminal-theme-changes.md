# Runtime terminal theme changes in the TUI task body

## Question

Why does task-body text keep its old foreground color when a terminal switches
between light and dark themes while `kanban-md tui` is running, and what is the
smallest robust fix?

## Method

- Reproduced the behavior deterministically through the light- and
  dark-background Markdown render paths.
- Inspected the implementation and history of task-body rendering, including
  the fix for task #249.
- Inspected the pinned Lip Gloss and Bubble Tea v1 source used by this project.
- Compared the available terminal-color APIs in Bubble Tea v1 with the newer
  Bubble Tea v2 API.

## Findings

- Task #249 changed the renderer from a permanently dark Glamour palette to a
  light or dark palette selected through `lipgloss.HasDarkBackground()`.
- The pinned Lip Gloss renderer guards background detection with `sync.Once`.
  Once queried, `HasDarkBackground()` returns the cached value for the rest of
  the process.
- Bubble Tea v1.3.10 deliberately triggers that query during package
  initialization, before the program acquires the terminal. It does not expose
  Bubble Tea v2's `RequestBackgroundColor` and `BackgroundColorMsg` runtime
  query mechanism.
- Glamour's light and dark styles apply fixed ANSI-256 foregrounds to the whole
  document: color 234 for a light terminal and color 252 for a dark terminal.
  Every plain body cell therefore keeps the palette chosen at startup.
- The rest of the detail view largely relies on the terminal's default
  foreground. Terminals update that default as their theme changes, which
  explains why the surrounding text adapts while the task body does not.

Bubble Tea v2 documents explicit foreground/background color queries, but a
v1-to-v2 migration changes core model, message, and view APIs and is much larger
than this body-text bug. See the official [Bubble Tea v2 overview](https://github.com/charmbracelet/bubbletea/discussions/1374)
and [upgrade guide](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md).

## Decision

Keep Glamour Markdown rendering and its initially selected light/dark palette
for semantic accents, but remove the palette's document-level foreground. Plain
task-body text then inherits the terminal's default foreground and follows
runtime theme changes without polling the terminal, parsing OSC responses, or
upgrading Bubble Tea.

The regression test renders the same plain body through both background paths
and verifies that neither fixed palette foreground is present and that both
paths produce identical output. Existing Markdown behavior tests continue to
cover headings, emphasis, inline code, lists, and wrapping.

## Live validation

A branch build was run in the locally installed cmux application. With task
#263 open in the detail view, the terminal default foreground and background
were changed through OSC 10/11 sequences after the TUI had started. The main
body remained readable for both light-to-dark and dark-to-light transitions
without restarting or reopening the task. The terminal colors and temporary
cmux workspace were restored and removed after the smoke test.

This validates the terminal-default color mechanism used by the fix. It does
not depend on cmux's macOS system-theme scheduler, so other terminals that
update their default colors at runtime should receive the same behavior.

## Scope note

Theme-specific Markdown accents such as link and syntax-highlight colors still
use the palette selected at process startup. The reported unreadability affects
the main body text; responding to every runtime palette change would require a
broader terminal-color event design, most naturally alongside a future Bubble
Tea v2 migration.
