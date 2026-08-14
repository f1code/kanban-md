# Issue #13 decision brief: optional task types as configurable guardrails

**Date:** 2026-08-14  
**Issue:** [#13 — Optional first-class `type` field, and configurable display fields](https://github.com/antopolskiy/kanban-md/issues/13)  
**Related issue:** [#11 — Show child tasks when viewing an epic](https://github.com/antopolskiy/kanban-md/issues/11)

## Executive recommendation

Issue #13 is trying to make a human-readable task category visible and useful without turning it into workflow automation. The concrete pain is that a board may use dependencies for work ordering, leaving every task at the same priority, while the human-scannable distinction (`milestone`, `epic`, `story`, `bug`) is buried among unrelated freeform tags.

The confirmed product principle should be:

> A task type is an optional, configured guardrail. Untyped tasks and boards with no configured types remain fully permissive. A configured type may change how its token is rendered and may reject a small set of explicitly selected mutations. It never makes a taxonomy mandatory and does not automate status, picking, claiming, roll-ups, or assignment.

For v1, use one optional scalar `type` on a task and a registry of sparse type definitions in config. Accept the issue's short string syntax as shorthand, but model each configured type as an object so it can independently carry an optional terminal color and two optional hierarchy capabilities: `may_be_child` and `may_have_children`. Omission means no restriction or override (effectively permissive in v1), never `false`.

```yaml
types:
  - name: milestone
    color: "63"
    may_be_child: false
  - name: epic
    color: "205"
  - name: story
    color: "39"
    may_have_children: false
  - name: bug
    color: "196"
    may_have_children: false

display:
  compact_fields: [status, type]
  tui_card_fields: [type]
```

The scalar shorthand remains useful for unconstrained types:

```yaml
types:
  - chore
  - spike
```

That means `chore` and `spike` form an enabled, ordered, enforced classification vocabulary, but add no color, display override, or hierarchy restriction.

## 1. The actual product problem

The issue is not primarily asking for another arbitrary metadata string. Tags already provide that. It is asking the tool to know which one category answers "what kind of work item is this?" so the category can:

- occupy a predictable display slot instead of being mixed into all tags;
- be filtered and grouped independently of size, milestone, ownership, and other tag dimensions;
- follow a board-defined display order rather than alphabetical tag order;
- carry presentation metadata, such as a stable color;
- optionally protect a few structural invariants selected by the board owner.

This remains a human-orchestration feature. The comments on issue #11 describe each child task as a separate agent run and parent/child as packaging for the human operator. They also explicitly argue that parent status should remain manual, with child progress displayed rather than derived. That supports keeping type out of `pick`, claim ownership, status transitions, and parent-status automation.

## 2. What issue #13 proposes, and what it only suggests

### Proposed in the issue

- Add an optional task `type` string.
- Add `types` to `config.yml`, initially shown as an ordered list of strings.
- Validate a non-empty task type against the configured list when that list is non-empty.
- Add `type` to JSON and to create/edit/list operations.
- Add exact type filtering and `--group-by type`, using config order for group order.
- Add configurable display fields, with the example default `display.compact_fields: [status, priority]`.
- Preserve existing boards by default; do not rewrite task files or migrate type-like tags automatically.
- Keep parent status, picking, ordering, and status rules independent from type.

### Suggested or explicitly left open

- Configured/closed types versus a freeform string.
- An arbitrary number of display fields versus a small fixed number of slots.
- Whether the TUI needs a type filter or type-oriented navigation in addition to showing type on cards.
- Whether compact output should be the only configurable surface or one setting should control compact CLI, `board`, and TUI cards.

### New product direction not present in the issue

The issue explicitly called type semantics a non-goal. The confirmed direction expands that scope: configured types may carry appearance metadata and bounded behavioral constraints, starting with whether a typed task may have a parent and whether other tasks may use it as a parent. This is a deliberate product decision beyond the issue text and should be acknowledged in the issue before implementation.

## 3. Current behavior and stale wording

### Current data and mutation model

- `internal/task.Task` has status, priority, tags, class, parent, and dependencies, but no type.
- Task JSON and YAML come from that same struct, so an optional `Type string` with `omitempty` would be additive.
- Create and edit converge on `internal/board` mutation functions. That is the correct shared enforcement boundary for CLI and TUI behavior.
- Parent existence and self-reference are already validated during create and edit. Issue #13's statement that existence and self-reference validation do not exist is stale. Cycles are still not detected.
- A parent relation and a dependency relation are separate concepts. Type hierarchy guardrails should apply only to `parent`, not `depends_on`.

### Current display surfaces

| Surface | Current task-level display | Consequence for type design |
|---|---|---|
| `list --compact`, compact `show`, compact `pick` details | `#ID [status/priority] title @claim (tags) due:...` | A compact field list maps naturally here, but output functions currently do not receive config. |
| `list` table | Fixed ID, status, priority, title, claim, tags, due columns | Type needs a conditional column or a future table-column setting; `compact_fields` should not silently reshape this table. |
| `show` table/detail | Status, priority, optional class, assignee, tags, due, timestamps | Show type only when set, near priority/class. |
| TUI card | Title, then priority plus tags, due, and age; status is already the column | Its current default is not `[status, priority]`. A single global compact-field default cannot preserve both compact CLI and TUI behavior. |
| TUI detail | Status, priority, optional class/assignee/tags/relations | Show type when set. |
| CLI `board` | Aggregate status/WIP/blocked/overdue and priority/class distributions, not task cards | `compact_fields` has no natural meaning here. `board --group-by type` is the clean v1 type surface. |
| TUI search/sort | Search by title or `#ID`; sort cycles priority, created, updated, title | Type filtering/navigation is additional scope, not already implied by card display. |
| CLI search/sort | Search title/body/tags; sort has no type | Exact type filter is required; search/sort behavior should be decided explicitly. |

The word `board` in issue #13 is therefore ambiguous. The current `board` command is an aggregate summary, while the interactive board is the TUI. Implementation should not assume one display configuration maps cleanly to both.

### How the broader schema compares with the original issue

The broader model is coherent, but not every part solves issue #13's original problem equally directly:

| Capability | Does it solve the original scanning problem? | Product fit and risk |
|---|---|---|
| First-class name, configured order, filter/group/sort | Yes, directly | Core issue #13 scope; low risk and additive. |
| Type color and card/detail display slots | Yes, directly | Strengthens the at-a-glance intent; low risk when opt-in and plain compact/JSON remain uncolored. |
| `may_have_children` / `may_be_child` | Not by itself | Adjacent guardrail enabled by the richer schema; bounded risk if omission stays permissive and explicit restrictions are enforced only on related writes. |
| Type-specific "show child progress" | It solves the other half of the human glance problem | This is derived board data and overlaps issue #11. Useful, but not necessary for first-class type and unsafe as the only way untyped parents reveal children. |
| Parent status inferred from children | No; it changes workflow rather than visibility | High risk and in tension with both issue #13's original non-goals and the concrete workflow described in issue #11. |

The schema therefore does not overshoot merely by being extensible. It overshoots when the first release makes type responsible for derived hierarchy displays or for writing workflow state. Static metadata and endpoint guardrails are a clean extension; child progress and status inference should have their own semantics and tests.

## 4. Recommended v1 product choices

### 4.1 One optional type, backed by a configured registry

- A task has zero or one type. Multi-category work remains the job of tags.
- An empty registry means the feature is off for normal writes. Under the explicit-config recommendation, a current-version file spells this as `types: []`; older configs may omit it until migration materializes the key. New tasks stay untyped; all parent/child shapes remain allowed; `--type VALUE` should tell the user to configure the type first.
- With a non-empty registry, new type values must match a configured name. Tags remain the freeform escape hatch.
- Do not add `defaults.type` or `require_type` in v1. Creating a task without `--type` must continue to produce an untyped, permissive task even on a board that has configured types.
- Keep the task field optional in YAML and JSON (`omitempty`) so untyped files and JSON are unchanged.

This is preferable to freeform type strings because the user wants type metadata, stable ordering, and guardrails. An empty registry should have one clear meaning rather than switching the same field into a second, weaker freeform mode.

### 4.2 Make config entries extensible from day one

Use a `TypeConfig` object with scalar shorthand, similar to the compatibility approach already used for status entries:

- `name` — required and unique;
- `color` — optional terminal color for the rendered type token;
- `may_be_child` — optional boolean; omitted means this type imposes no restriction on being a child;
- `may_have_children` — optional boolean; omitted means this type imposes no restriction on having children;
- future display properties — optional individually; omitted means inherit the surface's normal behavior.

Internally, behavioral booleans need tri-state representation (`*bool` or an equivalent optional value), not plain `bool`. The three states are:

| Raw value | Meaning in v1 |
|---|---|
| omitted / `nil` | No restriction and no override; effective allow. |
| `true` | Explicitly permits the capability; currently the same effective behavior as omission, but remains distinguishable in config. |
| `false` | Explicitly denies the capability and activates mutation validation. |

The distinction matters for sparse config, faithful YAML round-tripping, clear errors, and possible future inheritance. A Go zero value must never turn an omitted property into a restriction. Scalar shorthand creates a definition whose optional properties are all absent, not explicitly `false` or `true`.

The same sparse rule applies to display metadata. Omitted `color` means use the normal unstyled fallback; an omitted future `child_progress` setting means inherit global relation display. If an explicitly empty color is accepted, it should be normalized to no override; otherwise validation should reject the empty value with a precise message.

A non-empty propertyless list is meaningfully different from an empty list:

```yaml
types: [epic, story, bug]
```

- The registry is enabled and closed for new writes: those three values are valid and other `--type` values are rejected.
- Declaration order controls grouping and type sorting.
- Tasks may still omit `type`; untyped tasks remain fully permissive.
- The types add no color and no behavioral restriction.
- The original issue's classification/filter/group/display-slot problem is solved even without any properties.

By contrast, `types: []` leaves the registry off and new tasks untyped. Omission has the same effective meaning during old-version migration, but is a missing required key in a fully materialized current-version config.

The two capabilities express all minimal structural roles without encoding a hierarchy ladder:

| `may_be_child` | `may_have_children` | Typical role |
|---|---|---|
| false | true | top-level container, such as a milestone |
| true | true | nested container, such as an epic |
| true | false | leaf, such as a story or bug |
| false | false | standalone-only typed item |

Config order is display/group/sort order only. It must not imply that earlier types may parent later types.

### 4.3 Keep type-specific derived display out of the minimal schema

A concrete future type can reasonably express a detail-view preference:

```yaml
types:
  - name: epic
    color: "205"
    may_have_children: true
    may_be_child: false
    display:
      tui_detail:
        child_progress: summary
```

However, `child_progress` is not static type metadata like color. It requires loading other tasks, defining which children count, interpreting terminal statuses, and keeping detail scrolling correct in wide, narrow, and mouse TUI paths. That is the work already framed by issue #11.

The safest design is relation-first:

- any task with children can show a child list/progress, including untyped parents;
- type-specific display can later override presentation (`inherit`, `hidden`, `summary`, or `list`), but should default to `inherit`;
- direct children, archived-child handling, and terminal-status counting are global relation semantics, not reinvented per type;
- a restrictive `may_have_children: false` prevents new child links but does not hide legacy children that already exist.

This preserves the optional-type principle. If child visibility existed only under an `epic` type option, an untyped board would lose the very relation overview issue #11 is trying to add. The minimal type v1 should therefore reserve no nested display DSL beyond static color and the two global badge-field lists. Add type-specific detail options only after the relation-driven child display lands and its behavior is stable.

### 4.4 Enforce only prospective parent mutations

For a relationship `child.parent = parent.ID`, enforce both endpoints when their configured type supplies a policy:

- reject if the child's configured type has `may_be_child: false`;
- reject if the parent's configured type has `may_have_children: false`.

Validate at the shared `internal/board` mutation boundary after the proposed task state is known and before any file is written. The relevant cases are:

- creating a typed or untyped task with a parent;
- setting or changing a parent;
- changing a child's type while it has a parent;
- changing a parent's type while it already has children.

Clearing a parent must always be allowed. Clearing a type should also be allowed and deliberately returns that task to the permissive untyped behavior. That escape hatch follows the product principle; these are optional guardrails, not a mandatory schema.

Do not enforce policies on status moves, priority changes, claims, `pick`, dependencies, or child-progress roll-ups. Do not add allowed-parent-type matrices in v1.

### 4.5 Preserve and surface legacy/hand-edited data

Task parsing should never fail solely because a type is unknown to the current registry. Old or hand-edited task files may already contain `type`, and users must not be locked out of unrelated edits.

Recommended compatibility behavior:

- read and display an unknown type as its raw value, without configured color or restrictions;
- warn that it is not configured;
- allow `--clear-type` and unrelated edits;
- reject newly setting an unknown value while a registry is active;
- enforce hierarchy policy only for recognized configured types.

Changing config can make existing relationships inconsistent. Do not rewrite or detach tasks. Reads should remain available, and create/edit should reject only prospective relationship/type changes that violate policy. A later read-only board audit can list existing policy violations; auto-repair is out of scope.

### 4.6 Use explicit, surface-specific display slots

Do not use one `compact_fields` list for every surface. Preserve existing defaults with two small ordered lists:

```yaml
display:
  compact_fields: [status, priority] # exact current compact chip
  tui_card_fields: [priority]        # exact current TUI scalar badge
```

Existing TUI tags, due date, and age remain their current conditional annotations outside `tui_card_fields`. For v1, allow scalar badge fields such as `status`, `priority`, `type`, and `class`, with at most two entries per list. One or two entries are allowed; duplicates and unknown fields are config errors.

This gives the requesting board the intended result without changing defaults:

```yaml
display:
  compact_fields: [status, type]
  tui_card_fields: [type]
```

Additional display behavior:

- Human table list: add a TYPE column only when at least one returned task has a type; do not add an all-`--` column to untouched boards.
- Show/TUI detail: show Type only when set.
- Compact and JSON: never emit ANSI color. Compact is agent-oriented and must remain stable plain text.
- TUI and human table/group headings: apply the configured color only to the type token, with an unstyled fallback.
- CLI `board`: keep the existing summary by default; use `board --group-by type` rather than redefining compact fields as aggregate dimensions.

### 4.7 Complete the basic CLI lifecycle

Recommended v1 operations:

- `create --type NAME`;
- `edit --type NAME`;
- `edit --clear-type`, conflicting with `--type`;
- `list --type NAME[,NAME...]` with OR semantics;
- `list --untyped` for migration/backfill work;
- `list --group-by type` and `board --group-by type`;
- `list --sort type` using config order, with unknown values after configured ones and untyped last; `--reverse` reverses that order;
- include type in CLI `--search`, because that search already spans human text metadata.

Use `(untyped)` as the group label. When legacy unknown types exist, place configured groups first in config order, then unknown values alphabetically, then `(untyped)`.

Keep TUI search title/ID semantics and the TUI sort cycle unchanged in v1. Showing type on the card/detail is enough for the first interactive release. A TUI type selector/filter is useful follow-up work, but issue #13 does not require it and the current TUI create/edit wizard already omits several CLI fields such as parent and class.

### 4.8 Keep parent status manual; consider advisory inference only later

Persisted status inference is substantially riskier than child progress display:

- Status names and count are configurable. There is no universal `backlog`/`in-progress`/`done` mapping.
- The current `IsTerminalStatus` helper can identify the last non-archived status, but it cannot infer which arbitrary configured status means "started", "waiting", or "review".
- Valid states from the issue #11 discussion contradict simple derivation: a parent can be in progress between children, all children can be terminal while the parent still needs sign-off, and an empty parent has no obvious inferred state.
- Persisting an inferred move would write a second task file during a child mutation. The parent may be claimed by another agent, the inferred status may require a claim, and the move may violate WIP or class limits. The system would need an explicit transaction/rollback policy and audit-log semantics.
- Manual parent status is useful information in its own right. Replacing it with derived state removes the distinction between child completion and parent-level review.

The v1 answer should be child progress such as `9/23 done`, never an automatic parent move.

If users later ask for inference, start with a non-writing advisory on explicitly configured parent types. The smallest defensible semantics are:

```yaml
types:
  - name: epic
    status_suggestion:
      all_direct_children_terminal: done
```

- opt-in per configured type;
- direct, non-archived children only;
- zero included children produce no suggestion;
- "terminal" uses the board's existing terminal-status definition;
- the target (`done` above) must be an explicitly configured status;
- the result is displayed as a suggestion or drift warning and never written;
- manual status remains the source of truth.

Even `all children terminal -> done` is not universally correct because of parent sign-off, which is why it must be advisory. Any rule for an intermediate inferred status needs more explicit vocabulary: the user must define which child statuses count as active and which parent status to suggest. The tool should not guess from status order or names.

Persisted synchronization should remain deferred until there is evidence that advisory output is insufficient and the design answers: manual override, empty parents, direct versus recursive children, archived children, claim ownership, WIP/require-claim failures, concurrency, rollback, and audit logging.

## 5. Explicit generated configuration philosophy

### Current behavior is partly explicit and partly implicit

`NewDefault` constructs most active defaults and `init` serializes them, while version migrations fill several settings and immediately persist the migrated config. That is already close to explicit, versioned configuration.

It is not fully materialized today because many fields use `omitempty`. A normal generated board can omit active values such as:

- an empty `wip_limits` map (effective unlimited);
- `tui.hide_empty_columns: false`;
- `tui.narrow_threshold: 0` (effective automatic mode);
- `require_claim: false` and default-show duration on individual statuses;
- an empty board description and other optional metadata.

The runtime also contains fallback behavior with different strictness:

- empty TUI age thresholds silently use built-in thresholds;
- a zero title-line value has a helper fallback, although current config validation rejects it;
- empty claim timeout means no expiry;
- missing WIP limits mean unlimited;
- nil `show_duration` means show it;
- missing optional classes can disable class behavior.

As a result, reading the current YAML does not always reveal the complete active behavior. Validation catches some missing values because their Go zero value is invalid, but cannot distinguish an omitted `false`/`0` from an explicitly configured `false`/`0` for ordinary value fields.

### Product fit

A fully generated, inspectable board config fits kanban-md well:

- it is local, file-based infrastructure whose behavior should travel with the repository;
- humans and agents can inspect one file rather than know compiled defaults;
- identical repositories behave consistently across machines;
- config diffs make behavior changes during upgrades visible;
- the existing config version/migration system already provides the right evolution mechanism.

The cost is a longer first-run file and more migration churn. New users may mistake every exposed setting for something they must understand. Every new required field also needs a schema-version bump, migration, fixture, and compatibility test. Automatically materializing fields rewrites the YAML on upgrade, which can create noisy diffs and, as with current migrations, can reformat or discard hand-authored comments.

Portability improves across machines but becomes deliberately versioned across binaries: a newer migrated config remains rejected by an older binary. That is already the current safety behavior and should be documented as a one-way schema upgrade.

### Pragmatic recommendation

Adopt **explicit generated active settings**, not a rule that every optional object property must always appear.

1. `init` should materialize every board-level setting with one active value, including false/zero/empty values whose semantics matter. For the type feature that means at least `types: []`, `display.compact_fields: [status, priority]`, and `display.tui_card_fields: [priority]`.
2. A config declaring the current schema version should fail when a required active setting is absent. Do not silently reconstruct it from compiled defaults.
3. An older schema version should continue to load through migrations. Each migration supplies and persists the new active setting, so old boards upgrade rather than break.
4. Optional metadata and sparse overrides remain optional. An empty description, absent type color, and omitted type policy are not missing required settings; their absence is their documented value.
5. Runtime default constants remain useful for `init`, migrations, programmatic tests, and recovery messages, but should not silently fill missing required keys in a current-version file.
6. Add a future `config normalize`/materialize operation only if users need to convert hand-written current configs; do not auto-rewrite a valid current-version config on every load.

This requires presence-aware validation. Removing `omitempty` affects generated output but is not enough: ordinary Go `bool` and `int` fields cannot distinguish missing from explicit `false` or `0`. Validation would need either optional/pointer decode fields, custom unmarshal presence tracking, or a raw YAML key-presence pass keyed by schema version before decoding into runtime values.

The required/optional boundary should be semantic:

| Kind of setting | Generated? | Missing in current version | Example |
|---|---|---|---|
| Board-level active behavior | Yes | Validation error | status list, claim timeout, TUI hide-empty value, display field lists |
| Explicit empty feature registry | Yes | Validation error if the schema requires the key | `types: []`, `wip_limits: {}` |
| Optional descriptive metadata | Only when set, or explicit empty if desired for discoverability | No behavior error | board description |
| Sparse per-type metadata/policy | Only when set | Means no override/restriction | color, `may_be_child`, `may_have_children` |

There is a real tension between inspectability and sparse type semantics. Materializing every type boolean as `false` would accidentally restrict every propertyless type; materializing every boolean as `true` would erase the distinction between explicit permission and no override. The correct explicit representation of a propertyless type is therefore just its name. The top-level registry and global display defaults are materialized; the type object's omissions are documented first-class values.

For discoverability, a generated config can include a concise commented example or README link showing optional type properties. The active instance should not be padded with every possible `null` override merely to advertise the schema.

### Upgrade and compatibility rules under this philosophy

- Adding a new required board-level setting bumps the config version and its migration writes the chosen value.
- Adding a new optional sparse property may still require the repository's mandated version bump, but migration must leave it absent rather than invent an override.
- Same-version missing required keys fail with a message naming the key and explaining how to restore or migrate it.
- Old version fixtures prove migrations materialize all required active settings.
- An init-config golden test proves new boards expose the complete active configuration.
- A sparse-type round-trip test proves omission, explicit true, and explicit false remain distinct.

Plain-language recommendation: make new and migrated boards self-describing at the board-setting level, but treat "not specified" as a legitimate, visible-in-documentation value for optional type properties. That preserves reproducibility without turning an extensible type object into a wall of misleading defaults.

## 6. Compatibility and migration

This feature changes config schema and therefore must follow the repository's compatibility process even though task frontmatter is additive:

1. Bump `CurrentVersion` from 11 to 12.
2. Add the v11-to-v12 migration. It should materialize `types: []` and the two current display field lists, then increment the version.
3. Add `internal/config/testdata/compat/v11/` and a compat test proving the migrated config has an explicit empty type registry and explicit unchanged display defaults.
4. Add an optional task `type` fixture and task compat assertion without renaming/removing existing YAML fields.
5. Keep sparse properties inside each type omitted unless explicitly used, while materializing the top-level registry and active global display settings.
6. Do not rewrite task files or convert tags automatically.
7. Document a dry-run backfill recipe. Copy a recognized tag into `type` first; do not remove the tag automatically. Users can remove the old tag after reviewing results.

Sparse-schema compatibility tests should additionally prove:

- scalar and `{name: ...}` propertyless entries decode to absent optional properties;
- omitted, explicit `true`, and explicit `false` remain distinguishable through load/save;
- only present properties are validated;
- a non-empty propertyless registry enforces allowed names and configured order without restricting parent relations;
- migrating an old config does not materialize `false` values for absent properties.

Exact-output compatibility needs explicit tests:

- untyped default `list --compact` remains byte-for-byte `#ID [status/priority] ...`;
- current TUI golden files remain unchanged with the explicitly generated/migrated default display values;
- old JSON for an untyped task does not gain a non-empty or mandatory field;
- table output for a board with no types does not gain a blank TYPE column.

## 7. Meaningful decisions still open

The following should be confirmed before implementation:

1. **Scope of issue #13:** update it to include type metadata and parent/child guardrails, or keep #13 classification/display-only and file a linked follow-up for policies. Recommendation: design the object schema now and split implementation into reviewable PRs if desired; do not ship a string-only config that immediately needs migration.
2. **Empty registry:** does `types: []` turn the feature off, or make type freeform? Recommendation: off for new writes; tags are the freeform mechanism.
3. **Unknown existing values:** fatal, warning, or accepted silently? Recommendation: warning, readable/clearable, unstyled and unconstrained.
4. **Policy names:** `may_be_child`/`may_have_children`, `allow_parent`/`allow_children`, or negative flags such as `root_only`/`leaf`. Recommendation: `may_be_child`/`may_have_children`; the relationship endpoints are explicit. Omitted means no restriction, while explicit `false` denies the capability.
5. **Policy enforcement after config edits:** immediately block all commands, warn only, or enforce only prospective related mutations? Recommendation: never block reads or unrelated edits; warn about existing violations and enforce prospective relationship/type changes.
6. **Untyping as an escape hatch:** may clearing type remove its guardrails while relations remain? Recommendation: yes; otherwise types are not genuinely optional. Strict boards can consider `require_type` later.
7. **Cycles:** should this work add cycle detection? Recommendation: no. Current existence/self checks remain; cycle validation is a separate hierarchy-integrity feature.
8. **Pair-specific hierarchy:** should an epic accept stories but reject bugs, for example? Recommendation: defer. Two endpoint capabilities solve the stated need without an allowed-type matrix.
9. **Display configuration:** one list everywhere or surface-specific lists? Recommendation: separate compact and TUI card lists with unchanged defaults.
10. **Display-list size:** arbitrary or bounded? Recommendation: an ordered list of zero-to-two scalar badge fields; reject larger lists rather than allowing unreadable cards.
11. **Human table behavior:** always show TYPE, conditionally show it, or add table configuration? Recommendation: conditional column when returned data contains a type; table customization later.
12. **CLI board behavior:** add an automatic type distribution or rely on grouping? Recommendation: `board --group-by type` in v1; no new default section.
13. **TUI authoring/filtering:** add a type step and type filter now? Recommendation: display/detail only in v1, then a `None`-capable selector and filter if real use shows the need.
14. **Search:** should broad search match type? Recommendation: yes in CLI; leave TUI title/ID search unchanged for v1.
15. **Sort and untyped placement:** Recommendation: config order, then legacy unknown alphabetical values, then untyped; reverse flips the whole order.
16. **Color vocabulary:** terminal 0-255 strings only, or also hex colors? Recommendation: reuse Lip Gloss-compatible strings already familiar from TUI age thresholds, document accepted examples, and fall back unstyled if a terminal cannot represent the color.
17. **Default/required type:** Recommendation: neither in v1. Both can be added later without weakening the permissive default.
18. **Type-specific child display:** should `epic` be the only type that shows progress? Recommendation: no. Child discovery/progress is relation-driven for typed and untyped tasks; a later type option may override only its presentation.
19. **Progress shape:** count, direct-child list, or recursive tree? Recommendation: issue #11's direct children plus terminal/total count; no recursive tree in the type feature.
20. **Status inference:** computed hint or persisted automation? Recommendation: keep status manual in v1; if demand appears, trial an opt-in advisory `all_direct_children_terminal` suggestion before considering writes.
21. **Explicit-config boundary:** should all board-level active settings be present? Recommendation: yes for new/current-version configs; optional object overrides remain sparse.
22. **Missing current-version keys:** fallback or error? Recommendation: error for required active keys, with migrations supplying them only for older schema versions.
23. **Discoverability for sparse properties:** materialize null/default values or document them? Recommendation: keep instances sparse and provide a concise commented example/reference; do not turn omission into an artificial value.

## 8. What to defer

- Status derivation or cascades from child progress.
- Type-specific statuses, WIP limits, claim requirements, pick priority, assignee rules, or agent routing.
- Parent/child claim inheritance or requiring one agent for a subtree.
- Allowed parent/child type matrices and hierarchy depth rules.
- Type-driven dependency semantics.
- Mandatory types or a default type.
- Cycle detection and ancestor validation.
- Automatic tag-to-type migration or tag removal.
- TUI type tabs, swimlanes, keyboard navigation, bulk retagging, or type-specific views.
- Type-specific child-progress presentation until the relation-driven issue #11 behavior is settled.
- Any persisted inferred-status or parent-status synchronization.
- Automatic type distribution in the default board summary.
- Arbitrary card templates or unbounded display field lists.
- A full board-policy repair tool. A read-only audit is the appropriate first follow-up.

## 9. Recommended staged path

1. **Schema and visibility:** add the optional task field, object registry with scalar shorthand, color, set/clear/filter/group/sort, JSON, and opt-in compact/TUI badge fields. No derived data and no behavior constraints yet. Untyped output remains exact.
2. **Bounded hierarchy guardrails:** add sparse `may_be_child` and `may_have_children` properties. Omitted means no restriction; explicit `false` activates prospective validation. This can be a second PR in the same feature release so schema/display review is separable from mutation policy review.
3. **Relation-driven child progress:** complete issue #11 for any task with children. Use direct children, deliberate archived handling, terminal/total count, and manual parent status. Do not require a type.
4. **Type-specific presentation overrides:** only after stage 3, consider `inherit`/`hidden`/`summary`/`list` detail preferences. These alter presentation, not the underlying relation or status.
5. **Advisory status experiments:** if requested, expose a non-writing suggestion for an explicitly configured rule such as all direct children terminal. Measure whether it helps before designing intermediate-status rules.
6. **Persisted status inference:** no commitment. Treat as a separate workflow-automation feature requiring its own concurrency, claim, WIP, rollback, and override design.

## 10. Minimal acceptance-test and demo plan

Use a realistic board whose priority is intentionally constant so type provides visible value:

```yaml
types:
  - name: milestone
    color: "63"
    may_be_child: false
  - name: epic
    color: "205"
  - name: story
    color: "39"
    may_have_children: false
  - name: bug
    color: "196"
    may_have_children: false
display:
  compact_fields: [status, type]
  tui_card_fields: [type]
```

Create sample data:

1. `Release 1.0` — milestone, top level, medium priority.
2. `Authentication` — epic, parent #1, medium priority.
3. `Build login form` — story, parent #2, medium priority.
4. `Fix token refresh race` — bug, parent #2, medium priority.
5. `Investigate flaky CI` — untyped, top level, medium priority.
6. `Legacy imported task` — manually carries an unknown historical type to test compatibility.

### Demo checks

- Default board fixture with no `types` produces exactly today's compact/table/TUI output and allows arbitrary parent shapes.
- A newly initialized board's config golden materializes every required board-level active value, including false/zero/empty feature settings.
- Removing a required key from a current-version fixture fails with a precise error; the equivalent older-version fixture migrates and persists that key.
- A propertyless `types: [milestone, epic, story, bug]` registry rejects unknown new type names, preserves its order, and still allows every parent/child shape.
- Omitted, explicit-true, and explicit-false hierarchy properties exercise their three distinct raw states; only false blocks the relevant mutation in v1.
- The configured sample shows `[status/type]` in compact output, colored type tokens in human/TUI views, and no ANSI in compact/JSON.
- `list --type story,bug`, `list --untyped`, `list --sort type`, `list --group-by type`, and `board --group-by type` produce predictable configured order and an `(untyped)` bucket.
- JSON contains `type` only for typed tasks and keeps untyped tasks compatible.
- Creating epic #2 under milestone #1 succeeds.
- Creating story #3 and bug #4 under epic #2 succeeds.
- Giving milestone #1 a parent fails because milestone disallows a parent.
- Creating any child under story #3 fails because story disallows children, even when the proposed child is untyped.
- Changing epic #2 to story fails while #2 has children.
- Clearing #3's parent succeeds.
- Clearing a type succeeds and returns the task to permissive behavior.
- An untyped parent with an untyped child remains valid.
- A recognized restrictive endpoint still protects its side of a relation when the other endpoint is untyped.
- Unknown legacy type remains readable and clearable; unrelated edits work; setting that unknown value anew is rejected.
- Moving, claiming, releasing, blocking, and picking typed tasks behave exactly as they do for untyped tasks.
- Parent progress remains display-only and manual; no child completion changes a parent file.
- Wide and narrow TUI snapshots prove that one/two configured badge fields truncate safely and do not change card height unexpectedly.

### Acceptance threshold

The feature is ready for a user trial when the user can glance at the mixed sample board and distinguish milestone/epic/story/bug without reading tags, can opt out per task by leaving it untyped, receives a clear error for the two configured hierarchy violations, and observes no change at all in pick/claim/status behavior or on a board with no configured types.

## Bottom line

The cleanest v1 is not a freeform `type` string and not a new workflow engine. It is a small optional schema:

- one optional task value;
- an ordered registry with scalar shorthand;
- optional color;
- two independently optional, tri-state parent capabilities;
- a complete CLI set/clear/filter/group/sort lifecycle;
- surface-specific display slots with unchanged defaults;
- prospective mutation validation only;
- no automation.

That solves issue #13's scanning problem, creates room for principled metadata, and keeps the default Kanban model as flexible as it is today.
