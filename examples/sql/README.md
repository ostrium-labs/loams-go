# sql

Loads the same sample collection, then runs read-only SQL (aggregate and filter) and prints the rows. Set `LOAMS_SQL` to run your own statement, for example one using the `rrf`, `vector_search` or `text_search` table functions. DDL and writes are rejected by the engine.

```console
go run ./examples/sql
LOAMS_SQL="SELECT count(*) FROM docs" go run ./examples/sql
```
