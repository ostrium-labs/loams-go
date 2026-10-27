# quickstart

Uses the published Go SDK modules to ensure the namespace and `kb` collection exist, upsert three documents, run hybrid and vector searches, execute parameterized SQL, append and subscribe to a stream event, and optionally query the first existing Graph if that experimental feature is enabled. It does not drop or reset existing data.

```console
go run ./examples/quickstart
```

Graph is feature-detected; create a graph in the examples namespace to run the optional read-only GQL query.
