# searchsvc

An HTTP service using the Go SDK's collection handle and hybrid search module over `kb` from the quickstart. `ADDR` sets the listen address (default `127.0.0.1:8090`). The query vector is a toy keyword mapping; replace `embed` with a real model.

```console
go run ./examples/searchsvc
curl 'http://127.0.0.1:8090/search?q=refund&tenant=a&limit=3'
curl http://127.0.0.1:8090/healthz
```
