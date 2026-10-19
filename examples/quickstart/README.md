# quickstart

Connects with the Go SDK (`GetInstance`), creates `examples-go/docs` (text, keyword and 3-d vector fields), upserts four documents and runs one hybrid query: vector kNN and BM25 fused with RRF, filtered to tenant `a`.

```console
go run ./examples/quickstart
```
