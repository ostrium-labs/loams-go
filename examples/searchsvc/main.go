// Command searchsvc is a small HTTP search service over a Loams collection.
//
//	GET /search?q=refund&tenant=a&limit=3   -> JSON hits
//	GET /healthz                            -> the engine's name and version (via the SDK)
//
// The query vector is a toy: a hash-free keyword embedding, so the example
// needs no model. Swap embed() for a real one.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	loams "loams.dev/go"

	"github.com/ostrium-labs/loams-go/internal/rest"
)

// embed maps text onto the 3 axes the sample documents use.
func embed(q string) [3]float64 {
	q = strings.ToLower(q)
	v := [3]float64{}
	if strings.Contains(q, "refund") || strings.Contains(q, "damaged") {
		v[0] = 1
	}
	if strings.Contains(q, "gift") || strings.Contains(q, "card") {
		v[1] = 1
	}
	if strings.Contains(q, "return") || strings.Contains(q, "window") {
		v[2] = 1
	}
	if v == [3]float64{} {
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
	sdk, err := loams.New(loams.Options{Endpoint: rest.Endpoint(), Auth: loams.APIKey(os.Getenv("LOAMS_API_KEY"))})
	if err != nil {
		return err
	}
	defer sdk.Close()
	api := rest.New()
	setup, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := api.Setup(setup, "docs"); err != nil {
		return err
	}

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
		// A fresh client per request: rest.Client keeps the last consistency token.
		res, err := rest.New().HybridSearch(r.Context(), "docs", embed(q), q, r.URL.Query().Get("tenant"), limit)
		if err != nil {
			var he *rest.Error
			status := http.StatusBadGateway
			if errors.As(err, &he) && he.Status == 400 {
				status = http.StatusBadRequest
			}
			http.Error(w, err.Error(), status)
			return
		}
		writeJSON(w, res)
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
