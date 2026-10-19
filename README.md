# loams-go

Example apps for the [Loams](https://loams.dev) Go SDK (`loams.dev/go`, Go 1.25+).

Loams is an open-source engine for hybrid retrieval over your own bucket: vectors, BM25 and filters in one query, with SQL over the same collections. These examples show the Go SDK and the native REST API against a local engine.

| Example | What it shows |
|---|---|
| [quickstart](examples/quickstart) | Connect, create a collection, upsert documents, hybrid search (vector + BM25 + filter) |
| [sql](examples/sql) | Read-only SQL over a collection |
| [searchsvc](examples/searchsvc) | A small HTTP search service |

All examples read and write only the namespace `examples-go`.

## Note on the SDK surface

Go SDK 0.1.0 covers instance, live and tables. It has no collections, search or SQL module yet, so those calls go through `internal/rest`, a ~150 line helper over the native REST API. The SDK itself is used for connecting and for `GetInstance`. When the SDK grows those modules, the helper is the only thing to swap.

## Install

```console
go get loams.dev/go@v0.1.0
```

If `go get` cannot resolve `loams.dev/go@v0.1.0` yet (the module lives in `sdks/go` of ostrium-labs/loams and needs Go 1.25 plus the published tag), point at a checkout of the monorepo with a `go.work`:

```console
git clone https://github.com/ostrium-labs/loams ../loams
printf 'go 1.25.0\n\nuse .\n\nreplace loams.dev/go => ../loams/sdks/go\n' > go.work
```

## Run

Start an engine (from a Loams checkout, `cargo run -p loams -- dev`, or the compose stack), then:

```console
export LOAMS_ENDPOINT=http://127.0.0.1:8080   # your engine; `loams dev` needs no key
export LOAMS_API_KEY=...                      # only if auth is on
go run ./examples/quickstart
go run ./examples/sql
go run ./examples/searchsvc
```

Docs: <https://loams.dev/docs>. What ships in 0.1.0: [docs/release/preview.md](https://github.com/ostrium-labs/loams/blob/main/docs/release/preview.md).

## License

Apache-2.0.
