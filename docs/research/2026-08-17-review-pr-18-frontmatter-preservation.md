---
relationships:
  references:
    - https://github.com/antopolskiy/kanban-md/issues/16
    - https://github.com/antopolskiy/kanban-md/pull/18
---

# Review of PR 18: extra front-matter preservation

## Conclusion

PR 18 fixes a real data-loss problem: task mutations currently discard YAML
front-matter properties that are not represented by `task.Task`. The pull
request is large because it chooses a stronger contract than the issue needs.
It preserves not only the decoded values of unknown properties, but also YAML
presentation and graph structure: key order, comments, explicit tags, styles,
anchors, and aliases.

That stronger contract turns a small metadata-retention feature into a
syntax-aware merge algorithm. Under that contract, the current implementation
also has a blocking correctness gap: an alias can silently bind to a different
value after a canonical field is removed when an earlier anchor uses the same
name.

For the integration use case described in issue 16, semantic preservation of
unknown scalars, sequences, and mappings is the better scope unless maintaining
hand-authored YAML presentation is an explicit product requirement.

## What the pull request does

The implementation keeps two representations of a task:

1. the existing typed `Task` fields, which remain authoritative for kanban-md;
2. the original `yaml.Node` mapping, retained privately so unknown nodes can be
   merged back during writes.

On write, it encodes the typed task into a fresh canonical YAML mapping, merges
those canonical values into the original node tree, and re-parses the generated
YAML before touching the file. All mutation surfaces benefit because they
ultimately call the shared task writer.

The PR adds 1,565 lines across 13 files:

- 302 production-code lines;
- 1,001 test and fixture lines; and
- 262 README and research-documentation lines.

The test volume is therefore not evidence of a 1,500-line runtime feature. It
is evidence of the large behavioral promise the runtime code is making.

## Why the strong contract is complicated

### Typed data and source syntax must be reconciled

kanban-md must overwrite its own fields from typed Go values while retaining
unknown fields as source nodes. A plain untyped map would be easier, but would
weaken the typed domain model throughout the application.

### Anchors make key order semantic

An alias must follow the anchor to which it refers. Sorting unknown keys can
turn valid YAML into invalid YAML by moving an alias before its anchor. This is
why the PR retains the ordered node tree instead of using an inline node map.

### Canonical fields can own anchors used by unknown fields

Changing a canonical value should retain its anchor so an unknown alias keeps
working. Removing that canonical value can orphan the alias, so the writer must
detect the problem and refuse the mutation without changing the task file.

### Sequence items have no stable identity

If a tag is removed or reordered, copying an anchor or comment by index can
attach it to the wrong item. The PR therefore matches unique items by semantic
value, handles duplicate values conservatively, and carries comments from
removed items elsewhere.

### Comments belong to YAML nodes, not abstract fields

Removing a canonical field also removes the nodes that hold its comments. To
honor the presentation-preservation contract, the merge code must collect and
reattach those comments.

These rules explain the evolution from the initial implementation (133 lines
of merge code and 284 core tests) to the final implementation (288 lines of
merge code and 690 core tests), plus path-level tests. Most follow-up commits
addressed nested anchors, alias retargeting, removed-field comments, ambiguous
sequences, and cross-field alias authority.

## The contributor's suggested alternative

The contributor's PR comment suggests preserving only values and collections.
That is not yet a complete implementation proposal, but the implied contract is
clear:

- preserve unknown scalar values, lists, and mappings;
- allow re-encoding to normalize ordering and formatting;
- do not promise to retain comments, scalar style, explicit tags, anchors, or
  alias identity; and
- keep the extra properties out of table, compact, and JSON command output.

One likely implementation is to decode unknown properties into generic
semantic values and re-encode them alongside the typed canonical fields.
Aliases would be resolved to values, so key sorting could not orphan or retarget
them. This removes most presentation-copying, sequence-identity, comment
transfer, and alias-safety code.

The tradeoff is visible normalization: hand-authored comments and YAML reuse
constructs may disappear or expand, although integration-owned metadata values
remain intact. That is probably acceptable for the use case in issue 16, whose
goal is reliable storage for external-tool fields rather than lossless YAML
editing.

## Blocking review finding: duplicate anchor names can retarget an alias

YAML permits an anchor name to be reused; an alias refers to the most recent
preceding definition. The PR validates the generated YAML by parsing it, which
catches a missing anchor but cannot detect a syntactically valid change in
binding.

Reproduction against PR head `572590b`:

```yaml
id: &shared 1
title: Generic sample
status: todo
priority: medium
created: 2026-08-12T10:00:00Z
updated: 2026-08-12T10:00:00Z
estimate: &shared 4h
custom_copy: *shared
```

After reading the task, clearing `Estimate`, and calling `task.Write`, the write
succeeds. The `estimate` property is removed and `custom_copy` silently changes
from the string `"4h"` to the integer `1`, because `*shared` now binds to the
earlier `id` anchor. This violates the PR's stated promise to refuse writes that
ambiguously retarget aliases.

If the strong preservation contract is retained, add this as a regression test
and either compare alias target identity before and after the merge or reject
ambiguous duplicate anchor names. Merely re-parsing the generated YAML is not a
sufficient semantic validation.

## Verification

Against the PR head:

- `go test ./... -count=1` passed;
- compatibility tests passed;
- `go test -race ./internal/task -count=1` passed;
- `golangci-lint run --new-from-rev=<PR base> ./...` passed; and
- the duplicate-anchor adversarial test failed as described above.

The PR has no submitted reviews or inline review threads and remains a draft as
of 2026-08-17.

## Recommendation

Ask for the scope decision before reviewing individual implementation details:

1. If the product requirement is “external metadata values survive,” accept the
   contributor's simpler semantic-only direction and explicitly document that
   YAML formatting, comments, tags, anchors, aliases, and key order may be
   normalized.
2. If the requirement is “kanban-md behaves like a presentation-preserving YAML
   editor,” retain the node-overlay architecture, fix the duplicate-anchor
   retargeting bug, and treat the expanded tests and future maintenance burden
   as necessary rather than accidental complexity.

The first option is recommended for kanban-md's stated integration use case.
