---
relationships:
  references:
    - https://github.com/antopolskiy/kanban-md/issues/16
    - https://github.com/antopolskiy/kanban-md/pull/18
    - https://github.com/yaml/go-yaml
    - https://github.com/goccy/go-yaml
    - https://yaml.dev/doc/ruamel.yaml/overview/
    - https://github.com/mikefarah/yq
---

# YAML library boundary for extra task properties

## Conclusion

kanban-md is not reimplementing YAML parsing in PR 18. The existing
`go.yaml.in/yaml/v3` dependency already parses and emits YAML. The pull request
implements a different and substantially harder responsibility: a
presentation-preserving document editor that overlays a typed task model onto
an arbitrary YAML syntax tree.

Replacing the parser can reduce some mechanics, but no library can eliminate
the product rules around ownership, deletion, alias binding, and conflicts
between canonical and unknown fields. The largest simplification comes from
changing the contract, not the library:

> Treat task front matter as application data, not as a losslessly editable
> YAML document. Preserve unknown JSON-like values, while allowing YAML syntax
> and presentation to normalize.

The current YAML dependency is sufficient for that design. No service or new
runtime is needed.

## What is already offloaded

The repository uses `go.yaml.in/yaml/v3` to:

- tokenize and parse YAML;
- resolve anchors and aliases during semantic decoding;
- decode canonical task fields into typed Go values;
- decode arbitrary mappings, lists, and scalar values; and
- encode Go values back to valid YAML.

The library's `yaml.Node` also exposes tags, styles, anchors, aliases, and three
comment positions. Its documentation explicitly says the node representation
offers detailed control but does not preserve the original textual
representation when re-encoded. PR 18 builds the merge and preservation policy
on top of that node API.

The YAML organization now maintains this implementation. It considers v3 a
frozen compatibility line and directs ongoing feature development to v4. A v4
migration may be worthwhile independently, but it does not remove the typed
overlay or round-trip policy problem.

## Semantic-only probe with the existing dependency

A local probe decoded this input into a struct with a typed `id` and `title`
plus an inline `map[string]any` for unknown properties:

```yaml
id: 1
title: Example
custom_anchor: &shared
  enabled: true
custom_alias: *shared
custom_string: !!str 001
custom_tag: !External abc-123
custom_list:
  - one
  - two
```

The current dependency decoded all unknown values successfully. Re-encoding
produced ordinary YAML with these properties:

```yaml
custom_alias:
  enabled: true
custom_anchor:
  enabled: true
custom_list:
  - one
  - two
custom_string: "001"
custom_tag: abc-123
```

The values survived. The alias was expanded, the custom tag was removed, the
string was safely quoted, and keys and formatting were normalized. That is the
desired behavior under a semantic-only contract and avoids all alias-order and
retargeting logic.

## Recommended domain boundary

Define a persisted task as two kinds of data:

1. **Canonical fields** owned, typed, validated, and mutated by kanban-md.
2. **Extension values** carried through but otherwise ignored by kanban-md.

Extension values should support a deliberately small, JSON-compatible model:

- mappings with string keys;
- lists;
- strings;
- numbers;
- booleans; and
- null.

On read, decode canonical fields and collect all unknown keys into an extension
map. On write, encode canonical fields and the extension map together. Reject a
collision in favor of the canonical field. Keep extensions out of table,
compact, and JSON command output unless a future feature intentionally exposes
them.

Explicitly document that comments, key order, quoting style, block/flow style,
custom YAML tags, anchor names, aliases, and merge syntax are not part of the
persistence contract. Their resolved values survive; their YAML presentation
may not.

### An even cleaner schema

If compatibility with arbitrary top-level properties is not required, put
external metadata under one canonical namespace:

```yaml
extensions:
  notebook:
    session_record: abc-123
    attempts: 4
```

Then `Task` can have an ordinary field such as
`Extensions map[string]any` with `yaml:"extensions,omitempty"` and `json:"-"`.
This needs almost no custom serialization, avoids future name collisions, and
makes the ownership boundary visible in every task file. It is the strongest
domain simplification, although it asks integrations to use the namespace
instead of arbitrary top-level keys.

## Library and service options

| Option | What it helps with | What remains | Recommendation |
| --- | --- | --- | --- |
| Existing `go.yaml.in/yaml/v3` | Typed and generic semantic decode/encode | Classifying canonical versus extension keys | Use for the semantic-only design |
| `go.yaml.in/yaml/v4` | Actively developed successor and richer processing pipeline | Same ownership and round-trip policy | Evaluate separately; not required for this feature |
| `goccy/go-yaml` | Ordered maps, comment maps, AST/YAMLPath editing, anchor features, and advertised reversible transformations | Canonical/unknown merge rules, deletion semantics, alias identity, and regression coverage | Best Go spike if full presentation preservation is mandatory |
| `ruamel.yaml` | Mature Python round-trip editing of comments, order, styles, anchors, and merges | Python runtime/distribution plus edge cases when deleting keys or list entries | Technically capable, operationally wrong for a standalone Go CLI |
| `yq` | In-place YAML path updates and manipulation of comments, styles, tags, and anchors | External process/binary, mapping every task mutation to expressions, and documented preservation limitations | Do not add as a runtime dependency |
| Remote service | Could centralize parsing implementation | Network, privacy, latency, authentication, availability, version skew, and offline failure | Reject |

## Why a round-trip library is not a complete escape hatch

`goccy/go-yaml` is the most relevant Go candidate because it explicitly offers
comment maps, ordered maps, AST editing, YAML paths, and reversible
transformations. It may substantially reduce custom node-copying code. It still
cannot decide what should happen when:

- an unknown alias points to a canonical field that kanban-md removes;
- duplicate anchor names make a surviving alias change its target;
- a canonical sequence is reordered or contains duplicate values;
- comments belong to a field being deleted; or
- a future canonical field collides with a previously unknown property.

Those are application semantics. If full presentation preservation remains the
requirement, run a bounded prototype against these exact adversarial cases
before replacing the current dependency. Do not accept a library's general
“round-trip” label as proof of kanban-md's stronger contract.

`ruamel.yaml` is intentionally built for round-trip editing, but its own
documentation notes that preservation is normally reliable until structures
are severely altered, including key deletion and list-entry removal. Those are
ordinary kanban-md mutations. Calling Python locally or through a service would
exchange visible Go code for a more fragile distribution and runtime boundary.

`yq` is similarly capable but documents that formatting and comments are not
preserved in every scenario, and its anchor/merge behavior has compatibility
flags and known edge cases. It is useful as a user tool, not as an internal
serialization service for kanban-md.

## Decision

Do not introduce a YAML service, Python runtime, or `yq` subprocess. Do not
switch YAML libraries merely to preserve a requirement that is outside the
product's core domain.

Use the existing Go YAML library and implement semantic extension values. If
the project can choose the file schema, prefer a single `extensions:` mapping;
otherwise preserve arbitrary unknown top-level keys as generic values. This
keeps the domain focused on task management while YAML remains an encoding
format rather than a document-editing subsystem.
