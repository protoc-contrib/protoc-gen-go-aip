# protoc-gen-go-aip

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.md)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)

A [protoc](https://protobuf.dev) plugin that emits Go helpers for [Google
AIP](https://aip.dev) resource patterns and List-RPC query handling. It is
a unification of two earlier `protoc-contrib` plugins, with selected
additions adopted from [`go.einride.tech/aip`](https://github.com/einride/aip-go).

For each `.proto` it emits two companion files in the same Go package:

- **`*_aip.pb.resource.go`** — resource-name parsers and helpers driven
  by `google.api.resource` and `google.api.resource_reference`.
- **`*_aip.pb.query.go`** — AIP-160 filter / AIP-132 ordering /
  AIP-158 pagination helpers driven by `(protoc_contrib.aip.filterable)`,
  `(protoc_contrib.aip.orderable)`, and `(protoc_contrib.aip.column)`
  field options.

> **⚠ Binary-name collision.** This plugin's binary is `protoc-gen-go-aip`,
> the same name used by the upstream einride plugin under
> [`go.einride.tech/aip/cmd/protoc-gen-go-aip`](https://pkg.go.dev/go.einride.tech/aip/cmd/protoc-gen-go-aip).
> The two generate **different** APIs and are not interchangeable — install
> only one. If you need einride's resource-name layout (with its
> `MarshalString` / `UnmarshalString` interface and string-only segments),
> use einride's plugin. If you need this plugin's UUID-typed segments,
> cross-package references, and the AIP query helpers, use this one.

## Features

### Resource-name pass

- **Single-pattern resources** — emits `type <Type>Name struct { ... }`
  with one field per `{variable}` segment, plus `Parse<Type>Name`,
  `ParseFull<Type>Name`, `String()`, `FullName()`, `MarshalText` /
  `UnmarshalText`.
- **Multi-pattern resources** — emits a sealed `<Type>Name` interface
  and one struct per pattern named after its parent
  (e.g. `PublisherBookName`, `AuthorBookName`), plus a polymorphic
  `Parse<Type>Name` that tries each pattern in declaration order.
- **`Parent()` navigation** — child resources get a `Parent()` method
  returning the matched parent's generated type. Each pattern of a
  multi-pattern resource returns its own parent type.
- **Parent constructors** — the parent struct gains a method named after
  the child type that builds the child by inheriting parent fields and
  taking only the child-only segments as arguments
  (e.g. `parent.ProjectThingName(thingID)`).
- **Resource references** — every field annotated with
  `google.api.resource_reference` (including cross-package references)
  gains a `Parse<Field>()` method on the owning message that delegates to
  the referent's parser. Set the plugin option
  `allow_unresolved_refs=true` to skip references whose target type
  isn't in the compilation unit.
- **`Validate()` / `Type()` / `Pattern()` / `ContainsWildcard()`** —
  every generated struct exposes these AIP-122/159 helpers (adopted from
  einride's plugin).
- **File-level resources** — `google.api.resource_definition` at file
  scope emits parsers even without a backing message.
- **UUID-typed segments** — declare `(google.api.field_info).format = UUID4`
  on the `<var>_id` field of a `Create<Resource>Request` (AIP-133) to
  have the generated struct field typed as `uuid.UUID` and validated at
  parse time. The parent UUID consistency check is automatic across the
  pattern tree.

### Query pass

- **`ParseFilter()`** — emits a `filtering.Filter` parser using
  `filtering.Declarations` derived from `(protoc_contrib.aip.filterable)`
  fields.
- **`ParseOrderBy()`** — emits an `ordering.OrderBy` parser keyed off
  `(protoc_contrib.aip.orderable)` fields.
- **`ParsePageToken()`** — emits a `pagination.PageToken` parser when
  the `List<Resource>Request` has `page_token` and `page_size`.
- **`ParseQuery()`** — emits a single helper that runs all three at once
  and returns a typed `Query` struct.
- **`<Resource>Columns`** — a `map[string]string` projecting filterable
  / orderable fields to their backing DB column names, overridable via
  `(protoc_contrib.aip.column)`.

## Migration

Replace prior `protoc-gen-go-aip-resource` and `protoc-gen-go-aip-query`
plugin entries in `buf.gen.yaml` with a single `protoc-gen-go-aip`
entry. Output suffixes are unchanged (`_aip.pb.resource.go`,
`_aip.pb.query.go`), so downstream import paths don't churn.

## License

MIT — see [LICENSE.md](LICENSE.md).
