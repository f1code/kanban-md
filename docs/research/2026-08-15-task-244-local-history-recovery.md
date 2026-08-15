# Task 244 local history recovery

## Context

Task #244 implemented child roll-ups at `a2d933d` and then merged the current
`main` into its feature branch at `257ab72`. While task #249 was being fixed, a
shared-worktree branch race caused `main` to fast-forward through `257ab72` and
the #249 commit `3cdb671`. Commit `8c48c1a` then reverted #244 from `main` while
preserving #249.

Merging that `main` back into the #244 branch fast-forwarded the branch through
`8c48c1a`, so the prototype disappeared from its working tree even though its
commits and detached worktree remained intact.

## Evidence

- The #244 branch reflog recorded `257ab72` before the fast-forward to
  `8c48c1a`.
- `main` and the #244 branch both pointed to `8c48c1a` after that fast-forward.
- The detached worktree at `/Users/santop/Projects/kanban-md-task-244` remained
  clean at `a2d933d`, with the prototype source and demo board present.
- Neither `257ab72`, `3cdb671`, nor `8c48c1a` was contained in a remote branch.

## Recovery decision

Because the commits were local-only, rebuild a clean linear history instead of
adding a revert-of-revert:

1. Start from the last clean `main`, `ed9863a`.
2. Reapply the task-detail contrast fix as a standalone commit.
3. Apply the #244 merge tree relative to its `main` parent, producing one
   child-roll-up feature commit on the updated base.
4. Run the full verification pipeline before moving local branch references.

This keeps both user-facing changes while removing the accidental merge and
compensating revert from the rewritten history.
