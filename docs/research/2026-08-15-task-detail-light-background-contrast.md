# TUI task-detail contrast on light backgrounds

## Question

Was the pale task-body text on white terminal backgrounds always present, or
was it introduced by a later change?

## Method

- Located task-body rendering in `internal/tui/board.go`.
- Used `git blame` on `renderMarkdown` and searched history for the hard-coded
  `WithStandardStyle("dark")` call.
- Inspected the introducing commit and Glamour's bundled dark/light palettes.

## Findings

- Before commit `c8ab3793` (`feat: add markdown rendering to TUI task detail
  view`, 2026-02-11), task bodies were wrapped by Lip Gloss without an explicit
  foreground color, so the terminal's own default foreground was used.
- Commit `c8ab3793` introduced Glamour Markdown rendering and explicitly chose
  its `dark` palette.
- That hard-coded palette selection has not changed since its introduction.
- Glamour's dark palette renders document text with ANSI color 252, a pale gray
  intended for dark backgrounds. Its paired light palette uses ANSI color 234,
  a dark foreground intended for light backgrounds.

## Decision

Select Glamour's light or dark palette from Lip Gloss's detected terminal
background. Preserve the existing dark-background appearance while using the
light palette's darker task-body text on white/light terminals. Cover the light
path with a regression test that verifies ANSI color 234.
