# searchsvc

An HTTP service over the sample collection. `ADDR` sets the listen address (default `127.0.0.1:8090`). The query embedding is a toy keyword mapping so no model is needed; replace `embed` with a real one.

```console
go run ./examples/searchsvc
curl 'http://127.0.0.1:8090/search?q=refund&tenant=a&limit=3'
curl http://127.0.0.1:8090/healthz
```
