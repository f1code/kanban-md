# Markdown front matter interoperability conventions

**Date:** 2026-08-17
**Question:** Do Markdown applications conventionally support the full YAML language in front matter, including presentation details such as comments, anchors, aliases, tags, quoting, and formatting?

## Conclusion

No common cross-tool contract says that a Markdown application must preserve every valid YAML construct when it edits front matter. In practice, “YAML front matter” usually means a fenced metadata block whose values are parsed into the application's ordinary data model: scalars, arrays/lists, and maps/objects. Advanced YAML features can be accepted by a particular parser, but they are generally not portable application-level features.

The most important distinction is between readers and writers:

- Static-site generators and documentation builders usually read front matter and leave the source file untouched. They can parse YAML into a map without having to reconstruct its original spelling, comments, anchors, or aliases.
- Markdown editors and content-management tools write front matter back to disk. They commonly define supported field types or schemas and allow their serializer to normalize quoting, ordering, indentation, and syntax.

Therefore, it would be conventional for `kanban-md` to preserve additional front matter **values** while documenting that YAML presentation details and graph-level constructs may be normalized. The feature should not be described as lossless preservation of arbitrary YAML unless it is deliberately implemented and tested as such.

## Representative tools

| Tool | Application-level model | Does it normally rewrite Markdown? | Relevant convention |
| --- | --- | --- | --- |
| Obsidian | Text, lists, numbers, checkboxes, dates, date-times, and tags | Yes | Treats properties as typed data. Nested properties are not supported by the Properties UI. JSON front matter is accepted but converted back to YAML when saved. |
| Front Matter CMS | Schema-defined field types, including scalars, lists, taxonomy fields, nested fields, and blocks | Yes | The content model is explicit and serialization has configurable formatting conventions. |
| Decap CMS | Schema-defined widgets/fields persisted as front matter | Yes | The CMS owns the data model and serializes it to YAML front matter. |
| `python-frontmatter` | Python dictionary | Yes | The YAML handler loads metadata into a dictionary and exports the dictionary back to YAML. |
| Hugo | Booleans, integers, floats, strings, arrays, and maps | Generally no; it builds from source | Reserves standard fields and recommends placing custom fields under `params`. |
| Jekyll | YAML variables exposed to Liquid templates | Generally no; it builds from source | Requires valid YAML, but does not promise a lossless source round trip. |
| MkDocs | YAML metadata parsed to a Python dictionary | Generally no; it builds from source | Requires a top-level key/value collection and exposes parsed values, not YAML syntax. |
| Astro | YAML or TOML values, optionally checked by a content schema | Generally no; it builds from source | Treats front matter as content data and can validate it against application types. |

## Obsidian is a strong example of semantic rather than syntactic preservation

Obsidian's official documentation calls front matter entries “properties” and documents a deliberately small set of types: text, lists, numbers, checkboxes, dates, date-times, and tags. It says nested properties are not supported in the Properties UI, although they can be entered in source mode. It also says Markdown is intentionally unsupported inside properties because properties are intended to hold small, atomic pieces of information that are readable by both people and machines.

The clearest signal is format conversion: Obsidian accepts JSON inside the front matter delimiters, but interprets and saves it back as YAML. That is incompatible with a promise to preserve the original YAML/JSON presentation. Obsidian's contract is the resulting property data, not the exact serialization that represented it.

Source: [Obsidian properties documentation](https://obsidian.md/help/properties)

There are also community reports that Obsidian's front matter mutation API can change or remove quotes, comments, types, and formatting. This is secondary evidence rather than the primary documented contract, but it illustrates the consequences of an object-to-YAML rewrite.

Source: [Obsidian forum report about `processFrontMatter`](https://forum.obsidian.md/t/yaml-properties-api-processfrontmatter-removes-string-quotes-comments-types-formatting/65851)

## Editors and CMS tools typically establish a schema

Front Matter CMS documents a set of supported field types rather than presenting arbitrary YAML syntax as its domain. Its types include strings, numbers, dates, booleans, files/images, choices, lists, taxonomies, nested fields, field collections, and blocks. It also exposes output settings such as YAML array indentation and quote removal. Those formatting controls are evidence that the tool owns and normalizes serialization.

Sources: [Front Matter CMS field types](https://frontmatter.codes/docs/content-creation/fields), [Front Matter CMS settings](https://frontmatter.codes/docs/settings/overview)

Decap CMS similarly requires a content model made of fields and widgets. Those values are stored in front matter using a configured format such as YAML. Its public model is the CMS field schema, not every construct permitted by the YAML specification.

Sources: [Decap CMS configuration](https://decapcms.org/docs/configure-decap-cms/), [Decap CMS format options](https://decapcms.org/docs/configuration-options/), [Decap CMS repository](https://github.com/decaporg/decap-cms)

The `python-frontmatter` library makes the same boundary explicit at a lower level: its YAML handler parses metadata into a Python dictionary and exports a dictionary back to text. A plain dictionary cannot represent comments, scalar styles, original key spelling/order guarantees, anchors, or alias identity without a richer round-trip representation.

Source: [`python-frontmatter` handlers](https://python-frontmatter.readthedocs.io/en/latest/handlers.html)

## Read-only consumers avoid the hard problem rather than solve it

Jekyll, Hugo, MkDocs, and Astro mainly consume source files to produce a site. Their front matter documentation describes the data made available to templates or content code. Because these tools do not ordinarily rewrite the input Markdown, accepting valid YAML does not require reconstructing the original document later.

- Jekyll requires valid YAML between front matter delimiters and exposes custom variables to Liquid templates: [Jekyll front matter](https://jekyllrb.com/docs/front-matter/).
- MkDocs parses YAML-style metadata into a Python dictionary and requires the top level to be a key/value collection: [MkDocs metadata](https://www.mkdocs.org/user-guide/writing-your-docs/).
- Astro treats YAML or TOML front matter as custom content properties and can apply a schema: [Astro Markdown content](https://v5.docs.astro.build/en/guides/markdown-content/).
- Hugo documents front matter values as booleans, integers, floats, strings, arrays, and maps. It reserves built-in fields and recommends putting custom fields under `params`: [Hugo front matter](https://gohugo.io/content-management/front-matter/).

This is why it can be misleading to infer a lossless-editing requirement from what a static-site generator can read. Parsing and consuming a YAML document is much simpler than mutating a few known values and faithfully recreating everything else.

## What is portable in practice

The broadly interoperable subset is close to JSON's data model, with YAML's convenient scalar syntax:

- string, number, boolean, and null values;
- lists of values;
- string-keyed maps containing those values;
- sometimes application-recognized dates and date-times.

The following are valid YAML features but are not dependable cross-application front matter features:

- comments and their exact placement;
- quoting and block/flow scalar style;
- indentation and whitespace;
- anchors, aliases, and merge keys;
- custom YAML tags and application-specific types;
- duplicate keys;
- key order as semantic information;
- directives and multi-document streams.

Individual tools may accept some of these. Acceptance by a parser is different from an interoperability or round-trip guarantee.

## Implications for `kanban-md`

A substantially simpler and conventional contract would be:

1. `kanban-md` owns its documented task fields.
2. It preserves additional metadata as semantic scalars, lists, and string-keyed maps.
3. It may normalize YAML syntax whenever it writes a task.
4. It explicitly does not guarantee preservation of comments, anchors/aliases, custom tags, duplicate keys, ordering, quoting, or scalar style.
5. Unsupported values produce a clear validation error rather than being silently corrupted.

There are two reasonable shapes for additional metadata:

- Preserve unknown top-level values. This is convenient and closest to the current proposal, but risks future name collisions with new first-party fields.
- Add a namespaced map such as `extensions:` or `metadata:`. This creates a clean ownership boundary and follows Hugo's `params` precedent, but requires users and integrations to place custom data under that namespace.

If compatibility with existing task files is more important than a perfectly clean schema, `kanban-md` can support existing unknown top-level values while recommending the namespace for new integrations.

Suggested wording for the user-facing guarantee:

> `kanban-md` preserves additional front matter values composed of strings, numbers, booleans, nulls, lists, and string-keyed maps. When a task is updated, front matter is reserialized as YAML; comments, formatting, key order, anchors, aliases, custom tags, and other YAML presentation details may be normalized or removed.

This accurately describes the domain without implying that `kanban-md` is a general-purpose, lossless YAML editor.
