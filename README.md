# loams-go

These examples target **Loams SDK 0.1.2**. Installation and CI require the matching
0.1.2 packages and release tags to be published.

Example apps for the [Loams](https://loams.dev) Go SDK (`loams.dev/go`, Go 1.25+).

Loams is an open-source engine for hybrid retrieval over your own bucket: vectors, BM25 and filters in one query, with SQL over the same collections. These examples show the Go SDK and the native REST API against a local engine.

| Example | What it shows |
|---|---|
| [quickstart](examples/quickstart) | Connect, create a collection, upsert documents, hybrid search (vector + BM25 + filter) |
| [sql](examples/sql) | Read-only SQL over a collection |
| [searchsvc](examples/searchsvc) | A small HTTP search service |

All examples read and write only the namespace `examples-go`.

## Note on the SDK surface

The SDK provides namespaces, collections, documents, search, vector, graph, SQL and streams. These examples retain `internal/rest` for their native HTTP route demonstrations and use the SDK for instance discovery. See the [SDK notes example](https://github.com/ostrium-labs/loams/blob/v0.1.2/sdks/go/examples/notes/main.go) for builders, collection handles and SDK SQL.

## Install

```console
go get loams.dev/go@v0.1.2
```

For explicit local SDK development, point at a matching source checkout with a `go.work` (CI uses the published module):

```console
git clone --branch v0.1.2 --depth 1 https://github.com/ostrium-labs/loams ../loams
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

Docs: <https://loams.dev/docs>. What ships in 0.1.2: [docs/release/preview.md](https://github.com/ostrium-labs/loams/blob/main/docs/release/preview.md).

## License

Apache-2.0.
