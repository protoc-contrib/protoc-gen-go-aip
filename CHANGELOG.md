# Changelog

## [0.2.1](https://github.com/protoc-contrib/protoc-gen-go-aip/compare/v0.2.0...v0.2.1) (2026-10-10)


### Bug Fixes

* read fields through getters, skip repeated references, resolve types per package ([#57](https://github.com/protoc-contrib/protoc-gen-go-aip/issues/57)) ([e38cd16](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/e38cd16833dd4cbf86424d19befed14f0cf044d7))

## [0.2.0](https://github.com/protoc-contrib/protoc-gen-go-aip/compare/v0.1.3...v0.2.0) (2026-10-10)


### ⚠ BREAKING CHANGES

* filters using a CEL macro (has, all, exists, exists_one, map, filter) no longer compile in the generated FilterEnv.
* *_aip.pb.fieldmask.go is no longer generated, so Validate() disappears from update-request shaped messages. Delete stale *_aip.pb.fieldmask.go files when regenerating, and enforce update_mask paths with protovalidate's `field_mask.in` instead.
* the query pass no longer generates the Query type, Query.OrderByPaths, ParseQuery, ParseOrderBy, <Request>OrderByFields or ParsePageToken. A List request without a filter field gets no _aip.pb.query.go output.
* buf.build/protoc-contrib/protoc-gen-go-aip no longer publishes any protos. Remove the dep and the `import "protoc_contrib/aip/query.proto"` from consuming modules; the (protoc_contrib.aip.field_reference) annotation can be deleted outright.
* filter expressions are CEL, not AIP-160. `a = 1 AND b` becomes `a == 1 && b`. cel-go rejects the old syntax with a clear parse error rather than misreading it, so stale filters fail loudly.

### Features

* add the AIP-133 create-ID accessor and drop macros from the filter env ([#54](https://github.com/protoc-contrib/protoc-gen-go-aip/issues/54)) ([33ba6ab](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/33ba6abe10a069baa905959e0ea3573eb0fec939))
* drop the fieldmask Validate() pass ([#53](https://github.com/protoc-contrib/protoc-gen-go-aip/issues/53)) ([53b3225](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/53b3225ad8910260a18c727f530f901e0c97d0db))
* emit cel.dev/cel-go, drop the unused protoc_contrib/aip module ([#37](https://github.com/protoc-contrib/protoc-gen-go-aip/issues/37)) ([b608e69](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/b608e69724f5e719879e8f4eb38bcb6e80ff7126))
* generate against aip-go and cel-go, drop einride and field_reference ([34928f7](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/34928f728446957b1856311c663180c70678ffd9))
* generate only CEL filter helpers for List requests ([#51](https://github.com/protoc-contrib/protoc-gen-go-aip/issues/51)) ([5669257](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/5669257d2ae27337b9b882a99ecd9709611d6767))


### Bug Fixes

* **buf:** make fixture regeneration work ([eead33d](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/eead33d3199e5f20c57b9f84ca150f103ebae8c2))
* reject filters that do not evaluate to bool ([#52](https://github.com/protoc-contrib/protoc-gen-go-aip/issues/52)) ([78e4a37](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/78e4a3754b6401e3730cf01f2afde3c769658ed9)), closes [#46](https://github.com/protoc-contrib/protoc-gen-go-aip/issues/46)

## [0.1.3](https://github.com/protoc-contrib/protoc-gen-go-aip/compare/v0.1.2...v0.1.3) (2026-05-05)


### Bug Fixes

* **github:** correct action versions in update.yml ([a2f9925](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/a2f9925d1b11b49de17acfd73ce9a3c4d6ea5c84))

## [0.1.2](https://github.com/protoc-contrib/protoc-gen-go-aip/compare/v0.1.1...v0.1.2) (2026-04-26)


### Bug Fixes

* remove {Resource}Columns generation and protoc_contrib.aip.column option ([35790ba](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/35790ba3740203623b9c892ff2c2ced9d7bd357b))
* replace filterable/orderable annotations with field_reference ([c5457f5](https://github.com/protoc-contrib/protoc-gen-go-aip/commit/c5457f5d800640a9deb918e4113cded7b47c0956))

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
