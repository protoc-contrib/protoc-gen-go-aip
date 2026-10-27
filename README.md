# protoc-gen-go-aip

[![CI](https://github.com/protoc-contrib/protoc-gen-go-aip/actions/workflows/ci.yml/badge.svg)](https://github.com/protoc-contrib/protoc-gen-go-aip/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/protoc-contrib/protoc-gen-go-aip?include_prereleases)](https://github.com/protoc-contrib/protoc-gen-go-aip/releases)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.md)
[![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![protoc](https://img.shields.io/badge/protoc-compatible-blue)](https://protobuf.dev)

A [protoc](https://protobuf.dev) plugin that emits Go helpers for [Google
AIP](https://aip.dev) resource patterns and List-RPC query handling. It is
a unification of two earlier `protoc-contrib` plugins, with selected
additions adopted from [`go.einride.tech/aip`](https://github.com/einride/aip-go).

For each `.proto` it emits one `*_aip.pb.go` companion in the same Go
package, holding:

- **resource-name parsers and helpers** driven by `google.api.resource`
  and `google.api.resource_reference`;
- **AIP-160 CEL filter helpers** on List requests that carry a `filter`
  field. The resource is read off the List method's response, so nothing
  needs annotating.

A `.proto` with neither gets no companion.

> **⚠ Binary-name collision.** This plugin's binary is `protoc-gen-go-aip`,
> the same name used by the upstream einride plugin under
> [`go.einride.tech/aip/cmd/protoc-gen-go-aip`](https://pkg.go.dev/go.einride.tech/aip/cmd/protoc-gen-go-aip).
> The two generate **different** APIs and are not interchangeable — install
> only one. If you need einride's resource-name layout (with its
> `MarshalString` / `UnmarshalString` interface and string-only segments),
> use einride's plugin. If you need this plugin's UUID segments,
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
  `Parse<Type>Name` that tries each pattern in declaration order. The
  interface carries `Validate()`, `Type()`, `Pattern()` and
  `ContainsWildcard()` too, so a parsed name needs no type switch.
- **Pattern validation** — every `google.api.resource` pattern is compiled
  with aip-go's `CompileResourcePattern` before anything is generated, as
  protoc-gen-rust-aip does, so a schema that spells one wrong fails
  `buf generate` naming the message, the resource type and the pattern:
  an empty segment (`books//{book}`), path-template syntax
  (`{book=**}`), a stray brace, a literal that is not camelCase
  (`book_shelves`), a variable that is not snake_case (`{bookId}`), or a
  duplicate variable.
- **`Parent()` navigation** — child resources get a `Parent()` method
  returning the matched parent's generated type. Each pattern of a
  multi-pattern resource returns its own parent type.
- **A parent knows nothing of its children.** The dependency runs from the child
  to its parent (`Parent()`), never the reverse: a parent struct gains no method
  per child, so adding a child resource leaves the parent's generated API alone.
  Build a child's name from its parent's fields with a struct literal.
- **Resource references** — every singular string field annotated with
  `google.api.resource_reference` (including cross-package references)
  gains a `Parse<Field>()` method on the owning message that delegates to
  the referent's parser, reading the field through its getter so
  explicit-presence fields (proto2 `optional`, edition 2023) work. A
  repeated reference — AIP-231 Batch Get's `names` — and `type: "*"` are
  skipped. A type declared in several packages (v1 and v2 of one API)
  resolves to the referring package's own declaration, else to the only
  package declaring it; several candidates and none local is an error, and
  so is one package declaring the same type twice.
  Set the plugin option `allow_unresolved_refs=true` to skip references
  whose target type isn't in the compilation unit, or is ambiguous.
- **`Validate()` / `Type()` / `Pattern()` / `ContainsWildcard()`** —
  every generated struct exposes these AIP-122/159 helpers (adopted from
  einride's plugin).
- **File-level resources** — `google.api.resource_definition` at file
  scope emits parsers even without a backing message.
- **UUID segments** — declare `(google.api.field_info).format = UUID4`
  on the `<var>_id` field of a `Create<Resource>Request` (AIP-133) to
  have the segment validated as a UUID at parse time: a canonical lowercase
  one (AIP-122: an uppercase, braced, `urn:uuid:` or hyphen-less spelling is
  rejected, since it would format back as a different name), or the AIP-159
  wildcard `-`. The struct field stays a `string`, because a `uuid.UUID` cannot
  hold the wildcard, so a name keeps what it was given and `ContainsWildcard()`
  is the generic `n.X == "-"`. The parent UUID consistency check is automatic
  across the pattern tree.
  - **`Parse<Variable>ID()`** on the name returns the resource's *own* id as the
    standard library's `uuid.UUID` (so generated code needs Go 1.27). It fails
    for the wildcard, which is no id, and for a value that is not a UUID, which
    only a name built by hand holds. A parent's id is the parent's own:
    `name.Parent().Parse<Parent>ID()`.
  - **`Validate()`** requires every UUID segment to be a canonical, non-nil
    UUID, so the wildcard fails it: a write validates the name.
  - **`Format<Resource>Name(id uuid.UUID)`** takes a `uuid.UUID` and stores its string.
- **ID helpers for goverter** — a single-pattern resource whose own segment is
  UUID-typed also gets plain top-level functions, which goverter's `extend`
  directive can use and a method expression can't express:
  `Parse<Resource>ID(s string) (uuid.UUID, error)` parses the name, then its own
  id; `Parse<Resource><Parent>ID(s)` does the same for its parent's, when the
  parent is in the same package; and `Format<Resource>Name(id uuid.UUID) string`
  for a pattern with a single variable segment.
- **AIP-133 create IDs** — that same `Create<Resource>Request` gains a
  `Parse<Resource>ID()` method returning the ID the caller proposed, or
  `uuid.Nil()` when `<resource>_id` is empty: AIP-133 reads that as "the
  server assigns one", and which kind of UUID is the server's choice, so
  the accessor mints nothing —
  `if id == uuid.Nil() { id = uuid.NewV7() }` says which at the call site.
  Only for a single-pattern resource with a UUID-typed own ID, as in
  `protoc-gen-rust-aip`: a string ID has no validity rule the schema
  states, and a multi-pattern resource's create request does not say which
  pattern it creates under. It shares a name with the free function above
  but reads something different: the bare ID the request carries, not a
  resource name.

### Query pass

A request is a List request when all of these hold — the same rule
`protoc-gen-rust-aip` applies:

- a service **in the same `.proto` file** has a method, neither client- nor
  server-streaming, that takes it;
- that method's response has **exactly one** repeated message field (a map
  does not count), whose type is the resource;
- the request is a **top-level** message of that file — nested messages are
  not considered — with a singular `string filter` field;
- the resource has at least one field with a CEL type.

For each, the plugin emits:

- **`<Request>FilterEnv`** — a `*cel.Env` declaring every resource field
  that has a CEL type (strings, numbers, bools, enums as ints,
  `Timestamp`, `Duration`). Nested messages, repeated fields and maps are
  skipped. It has **no macros and no optional syntax**: `exists`, `has` and
  the like would expand to comprehensions and presence tests, and `a.?b` to
  an optional, none of which has a reading as a query — the same grammar
  `protoc-gen-rust-aip`'s environment accepts.
- **`ParseFilter()`** — compiles the `filter` expression against that
  environment and returns the checked `*cel.Ast`, or `(nil, nil)` when
  the filter is empty or blank. Pass the AST to a query layer such as
  [`pgxcel`](https://github.com/pgx-contrib/pgxcel), whose column map
  decides which fields a client may actually filter by.

Ordering (AIP-132) and page tokens (AIP-158) are not generated: both only
make sense against the column map, ordering and SQL that actually run,
which the `.proto` can't see, so they belong to the query layer.
`protoc-gen-rust-aip` generates nothing for them either.

## Example

Given this `books.proto`:

```proto
syntax = "proto3";

package books.v1;

import "google/api/resource.proto";
import "google/protobuf/timestamp.proto";

message Book {
  option (google.api.resource) = {
    type: "library.example.com/Book"
    pattern: "books/{book}"
  };

  string name = 1;
  string title = 2;
  string author = 3;
  google.protobuf.Timestamp create_time = 4;
}

message ListBooksRequest {
  int32 page_size = 1;
  string page_token = 2;
  string filter = 3;
}

message ListBooksResponse {
  repeated Book books = 1;
  string next_page_token = 2;
}

service Library {
  // ListBooksResponse's repeated Book makes this the List method for Book.
  rpc ListBooks(ListBooksRequest) returns (ListBooksResponse);
}
```

`buf generate` produces one companion, `books_aip.pb.go`. The highlights —
a resource-name parser:

```go
type BookName struct {
    BookID string
}

func ParseBookName(s string) (BookName, error)
func (n BookName)  String() string
func (n BookName)  Validate() error
func (x *Book)     ParseName() (BookName, error)
```

and a CEL filter parser:

```go
// Declares name, title, author and create_time.
var ListBooksFilterEnv *cel.Env

func (x *ListBooksRequest) ParseFilter() (*cel.Ast, error)
```

The methods on a message name the field they read: a failure is a
`*aip.FieldError` from [aip-go](https://github.com/protoc-contrib/aip-go), whose
`Field()` is the proto field (`name`, `parent`, `filter`) and whose `Unwrap` is
the parse failure. A handler that answers `InvalidArgument` takes the field from
the error rather than typing it again; the free `Parse<Resource>Name` functions
have no field and return their error bare. The generated code imports aip-go, so
a module that compiles it requires `github.com/protoc-contrib/aip-go`. The file
imports it as `aip_go`, the name protogen derives from the path's last element
(`aip-go`) and the one its doc comments use, since protogen has no way to pick
another; the package itself is named `aip`.

Call sites stay terse:

```go
name, err := ParseBookName("books/foo")        // BookName{BookID: "foo"}, nil

filter, err := req.ParseFilter()               // *cel.Ast checked against ListBooksFilterEnv; nil when blank
where, args, err := pgxcel.Where(filter, pgxcel.WithColumns(columns))
```

## Installation

```bash
go install github.com/protoc-contrib/protoc-gen-go-aip/cmd/protoc-gen-go-aip@latest
```

## Usage

### With buf

Add the plugin to your `buf.gen.yaml`:

```yaml
version: v2
plugins:
  - local: protoc-gen-go-aip
    out: .
    opt:
      - module=github.com/your-org/your-module
```

Then run:

```bash
buf generate
```

### With protoc

```bash
protoc \
  --go-aip_out=. \
  --go-aip_opt=module=github.com/your-org/your-module \
  -I proto/ \
  proto/example.proto
```

## Options

| Option                  | Default | Effect                                                                                                                  |
| ----------------------- | ------- | ----------------------------------------------------------------------------------------------------------------------- |
| `allow_unresolved_refs` | `false` | When `true`, `google.api.resource_reference` fields whose target type is not in the compilation unit, or is declared only in several other packages, are skipped silently rather than producing a codegen error. |

## Migration

Replace prior `protoc-gen-go-aip-resource` and `protoc-gen-go-aip-query`
plugin entries in `buf.gen.yaml` with a single `protoc-gen-go-aip`
entry. Their two companions, `_aip.pb.resource.go` and `_aip.pb.query.go`
— which this plugin also wrote up to 0.2.2 — are now one `_aip.pb.go`.
Delete the old files before regenerating (or set `clean: true` in
`buf.gen.yaml`), since a stale one redeclares everything the new one holds.
Import paths don't change — the output stays in the same Go package.

## Contributing

To set up a development environment with [Nix](https://nixos.org):

```bash
nix develop
go test ./...
```

Or, without Nix, ensure `go`, `protoc`, and `buf` are on your `PATH`.

## License

[MIT](LICENSE.md)
