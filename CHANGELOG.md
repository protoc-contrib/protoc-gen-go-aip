# Changelog

## [0.1.1](https://github.com/protoc-contrib/protoc-gen-go-aip/compare/v0.1.0...v0.1.1) (2026-04-25)


### Bug Fixes

* restructure buf.yaml to publish field options at protoc_contrib/aip/query.proto ([6c99809](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/6c998093c5d558fdceb2ca29955a8205d3d78416))

## [0.1.0](https://github.com/protoc-contrib/protoc-gen-go-aip/compare/v0.0.1...v0.1.0) (2026-04-25)


### Features

* add fieldmask pass for AIP-134 update request validation ([0906222](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/090622289def1630ee2f622a896f9277e5d90407))
* add protoc-gen-go-aip plugin for resource names and query helpers ([bdfed97](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/bdfed9791864f18560335e6c03bd610f0af91893))

## v0.1.0

Initial release. This plugin merges the previously separate
`protoc-gen-go-aip-resource` and `protoc-gen-go-aip-query` into a single
binary that emits two companion files per `.proto`: `*_aip.pb.resource.go`
and `*_aip.pb.query.go`.

### Resource pass

Carried over from `protoc-gen-go-aip-resource`:

- `<Type>Name` structs with `Parse<Type>Name` / `ParseFull<Type>Name` /
  `String()` / `FullName()` / `MarshalText` / `UnmarshalText`.
- `Parent()` navigation returning the parent resource's generated type.
- Multi-pattern sealed interface with parent-derived variant names
  (`PublisherBookName`, fallback `<Type>Name_<N>`).
- `Format<Type>Name` / `Parse<Type>ID` helpers for single-ID resources
  with typed segments.
- `Parse<Field>()` methods on messages with
  `google.api.resource_reference`, including cross-package targets.
- `allow_unresolved_refs` plugin option.
- UUID4-typed segments (struct field typed `uuid.UUID`, parse-time
  validation, automatic parent UUID consistency check).

Newly added (adopted from `go.einride.tech/aip/cmd/protoc-gen-go-aip`):

- `Validate()` — empty-segment and `/`-in-segment checks.
- `Type()` — returns the resource's `service/Type` string constant.
- `Pattern()` — returns the canonical pattern string constant.
- `ContainsWildcard()` — returns true if any segment equals `-`
  (AIP-159 wildcard).
- Parent constructors — the parent struct gains a method named after the
  child type that builds the child by inheriting parent fields and
  taking only the child-only segments as arguments (e.g.
  `parent.ProjectThingName(thingID)`).

### Query pass

Carried over from `protoc-gen-go-aip-query` at v0.7.0 design:

- `ParseFilter` / `ParseOrderBy` / `ParsePageToken` / `ParseQuery`
  helpers on `List<Resource>Request`.
- `<Resource>Columns` map combining filterable, orderable, and
  `column`-override fields.
- `<Resource>OrderByColumns` for column-driven projection.
- Driven by the `(protoc_contrib.aip.filterable)`,
  `(protoc_contrib.aip.orderable)`, and `(protoc_contrib.aip.column)`
  field options (carried over verbatim from the prior plugin).

### Binary

The CLI is `protoc-gen-go-aip`. Note this collides on `$PATH` with
`go.einride.tech/aip/cmd/protoc-gen-go-aip` — install only one.
