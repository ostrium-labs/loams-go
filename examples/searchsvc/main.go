// Command searchsvc serves GET /search over the Loams SDK collection handle
// and GET /healthz through the instance module.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	loams "loams.dev/go"
)

func embed(q string) []float32 {
	q = strings.ToLower(q)
	v := []float32{0, 0, 0}
	if strings.Contains(q, "refund") || strings.Contains(q, "damaged") {
		v[0] = 1
	}
	if strings.Contains(q, "gift") || strings.Contains(q, "card") {
		v[1] = 1
	}
	if strings.Contains(q, "return") || strings.Contains(q, "window") {
		v[2] = 1
	}
	if v[0] == 0 && v[1] == 0 && v[2] == 0 {
		v[0] = 0.1
	}
	return v
}

func main() {
	if err := run(); err != nil {
		log.Fatal("searchsvc: ", err)
	}
}

func run() error {
	namespace := os.Getenv("LOAMS_NAMESPACE")
	if namespace == "" {
		namespace = "examples-go"
	}
	endpoint := os.Getenv("LOAMS_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://127.0.0.1:8080"
	}
	var auth loams.TokenSource
	if key := os.Getenv("LOAMS_API_KEY"); key != "" {
		auth = loams.APIKey(key)
	}
	sdk, err := loams.New(loams.Options{Endpoint: endpoint, Auth: auth, Namespace: namespace, SessionConsistency: true})
	if err != nil {
		return err
	}
	defer sdk.Close()
	kb := sdk.Collection("kb", namespace)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		info, err := sdk.Instance().GetInstance(r.Context(), &loams.GetInstanceRequest{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		writeJSON(w, map[string]string{"name": info.GetName(), "version": info.GetServerVersion()})
	})
	mux.HandleFunc("GET /search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "" {
			http.Error(w, "q is required", http.StatusBadRequest)
			return
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 || limit > 50 {
			limit = 5
		}
		query := kb.Query().Text(q, loams.TextOptions{Fields: []string{"body"}}).
			Vector("embedding", embed(q), loams.VectorOptions{K: 20}).
			Fuse(loams.RRF(60)).Limit(uint32(limit))
		if tenant := r.URL.Query().Get("tenant"); tenant != "" {
			query = query.Filter(loams.Filter.Eq("tenant", tenant))
		}
		res, err := kb.Search(r.Context(), query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		hits := make([]map[string]any, 0, len(res.GetHits()))
		for _, hit := range res.GetHits() {
			hits = append(hits, map[string]any{
				"id": hit.GetPk().String(), "score": hit.GetScore(), "source": hit.GetSource().AsMap(),
			})
		}
		writeJSON(w, map[string]any{"hits": hits})
	})

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:8090"
	}
	fmt.Println("listening on", addr)
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	return srv.ListenAndServe()
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
