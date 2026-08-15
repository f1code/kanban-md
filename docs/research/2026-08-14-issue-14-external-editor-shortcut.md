# Issue #14: external editor shortcut

## Source

- GitHub issue: [Add keybinding to open ticket in `$EDITOR`](https://github.com/antopolskiy/kanban-md/issues/14)
- Requester: `david-haerer`
- Reviewed: 2026-08-14

## Request

The requester wants to jump from the TUI directly into their terminal editor with the selected task's Markdown file. In the issue discussion, the maintainer proposed the conventional `$VISUAL`-then-`$EDITOR` priority. The requester confirmed that behavior and pointed to lazygit's split between lowercase `e` for its built-in editor and uppercase `E` for the external editor.

## Decision

- Keep lowercase `e` unchanged for kanban-md's built-in four-step edit flow.
- Add uppercase `E` on the board view to open the selected task file.
- Resolve the editor from `$VISUAL`, falling back to `$EDITOR` and then to `vi` on `PATH` when the environment variables are empty or unset.
- Use Bubble Tea's external-process support so the TUI releases the terminal before the editor starts and restores it after the editor exits.
- Reload the task files and restore selection to the edited task after the external process completes.
- Show missing-editor and launch errors in the existing TUI error line.
- Document the shortcut in both the in-app help and README keyboard shortcut table.

## Scope

No configuration schema change is needed. The environment variables provide the requested configuration, so config migration and backward-compatibility fixtures are unaffected.

## Follow-up

On 2026-08-15, the fallback was extended to use `vi` when neither environment variable is set and `vi` is available on `PATH`. External-editor errors were also made transient: the next keyboard or mouse action dismisses the error line.
