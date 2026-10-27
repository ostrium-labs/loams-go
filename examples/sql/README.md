# sql

Runs parameterized read-only statements through `client.Sql().Records` over `kb` from the quickstart, including hybrid retrieval in SQL.

```console
go run ./examples/sql
LOAMS_SQL="SELECT count(*) AS n FROM kb" go run ./examples/sql
```
