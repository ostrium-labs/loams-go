// Package rest demonstrates the native HTTP JSON routes alongside the SDK.
// The SDK also provides app modules, builders, collection handles and SQL.
package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Namespace is the only namespace the examples touch.
const Namespace = "examples-go"

// Endpoint returns LOAMS_ENDPOINT or the default `loams dev` address.
func Endpoint() string {
	if e := os.Getenv("LOAMS_ENDPOINT"); e != "" {
		return strings.TrimRight(e, "/")
	}
	return "http://127.0.0.1:8080"
}

// Client posts JSON to the native API.
type Client struct {
	Base  string
	Key   string
	HTTP  *http.Client
	Token string // last consistency token, sent back for read-your-writes
}

// New builds a client from the environment (LOAMS_ENDPOINT, LOAMS_API_KEY).
func New() *Client {
	return &Client{Base: Endpoint(), Key: os.Getenv("LOAMS_API_KEY"), HTTP: &http.Client{Timeout: 30 * time.Second}}
}

// Error is a non-2xx answer.
type Error struct {
	Status int
	Body   string
}

func (e *Error) Error() string { return fmt.Sprintf("loams: HTTP %d: %s", e.Status, e.Body) }

// Post sends body as JSON and decodes the JSON answer into out (may be nil).
func (c *Client) Post(ctx context.Context, path string, body, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Base+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Key != "" {
		req.Header.Set("Authorization", "Bearer "+c.Key)
	}
	if c.Token != "" {
		req.Header.Set("Loams-Consistency-Token", c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if t := resp.Header.Get("Loams-Consistency-Token"); t != "" {
		c.Token = t
	}
	if resp.StatusCode/100 != 2 {
		return &Error{Status: resp.StatusCode, Body: strings.TrimSpace(string(data))}
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

// Delete removes a collection.
func (c *Client) Delete(ctx context.Context, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.Base+path, nil)
	if err != nil {
		return err
	}
	if c.Key != "" {
		req.Header.Set("Authorization", "Bearer "+c.Key)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 && resp.StatusCode != http.StatusNotFound {
		data, _ := io.ReadAll(resp.Body)
		return &Error{Status: resp.StatusCode, Body: string(data)}
	}
	return nil
}

// Map is shorthand for a JSON object.
type Map = map[string]any

// Setup makes the namespace and a fresh `docs` collection with a body text
// field, a tenant keyword, and a 3-dimensional cosine vector, then writes the
// sample documents. Re-running it resets the collection.
func (c *Client) Setup(ctx context.Context, collection string) error {
	if err := c.Post(ctx, "/v1/namespaces", Map{"name": Namespace}, nil); err != nil {
		var he *Error
		if !asError(err, &he) || (he.Status != 409 && he.Status != 400) {
			return err
		}
	}
	base := "/v1/namespaces/" + Namespace + "/collections"
	if err := c.Delete(ctx, base+"/"+collection); err != nil {
		return err
	}
	schema := Map{
		"fields": []Map{
			{"name": "body", "source_path": "body", "kind": Map{"text": Map{"analyzer": "standard", "positions": true}}, "indexed": true, "fast": false},
			{"name": "tenant", "source_path": "tenant", "kind": "keyword", "indexed": true, "fast": true},
			{"name": "n", "source_path": "n", "kind": "i64", "indexed": true, "fast": true},
		},
		"vectors":        []Map{{"name": "embedding", "dim": 3, "distance": "cosine"}},
		"sparse_vectors": []Map{},
		"dynamic":        "ignore",
		"max_fields":     1000,
	}
	if err := c.Post(ctx, base, Map{"name": collection, "schema": schema, "partitions": 1}, nil); err != nil {
		return err
	}
	up := func(id int, body, tenant string, v [3]float64) Map {
		return Map{"upsert": Map{"id": id, "source": Map{"body": body, "tenant": tenant, "n": id}, "vectors": Map{"embedding": v}}}
	}
	ops := []Map{
		up(1, "refund policy for damaged items", "a", [3]float64{1, 0, 0}),
		up(2, "shipping times and carriers", "a", [3]float64{0.9, 0.1, 0}),
		up(3, "refund window for returns", "b", [3]float64{0, 0, 1}),
		up(4, "gift card terms", "b", [3]float64{0.1, 0.9, 0}),
	}
	return c.Post(ctx, base+"/"+collection+"/documents", Map{"ops": ops}, nil)
}

// Hit is one search result.
type Hit struct {
	ID     any            `json:"id"`
	PK     any            `json:"pk"`
	Score  float64        `json:"score"`
	Source map[string]any `json:"source"`
}

// SearchResponse is the part of a search answer the examples read.
type SearchResponse struct {
	Hits []Hit `json:"hits"`
}

// HybridSearch runs one query: a vector kNN and a BM25 match fused with RRF,
// optionally narrowed by a tenant filter.
func (c *Client) HybridSearch(ctx context.Context, collection string, vec [3]float64, text, tenant string, limit int) (*SearchResponse, error) {
	req := Map{
		"collection": collection,
		"retrievers": []Map{
			{"vector": Map{"field": "embedding", "query": vec, "k": 10, "params": Map{"exact": true}}},
			{"text": Map{"query": Map{"match": Map{"field": "body", "text": text, "operator": "or"}}, "k": 10}},
		},
		"fusion": Map{"rrf": Map{"k": 60}},
		"limit":  limit,
		"select": Map{"source": "all", "vectors": []string{}, "fields": []string{}},
	}
	if tenant != "" {
		req["filter"] = Map{"term": Map{"field": "tenant", "value": tenant}}
	}
	var out SearchResponse
	if err := c.Post(ctx, "/v1/namespaces/"+Namespace+"/query", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SQL runs a read-only statement over the namespace's collections.
func (c *Client) SQL(ctx context.Context, query string) (cols []string, rows [][]any, err error) {
	var out struct {
		Columns []struct {
			Name string `json:"name"`
		} `json:"columns"`
		Rows [][]any `json:"rows"`
	}
	if err = c.Post(ctx, "/v1/namespaces/"+Namespace+"/sql", Map{"query": query}, &out); err != nil {
		return nil, nil, err
	}
	for _, col := range out.Columns {
		cols = append(cols, col.Name)
	}
	return cols, out.Rows, nil
}

func asError(err error, target **Error) bool {
	e, ok := err.(*Error)
	if ok {
		*target = e
	}
	return ok
}

// Title returns a hit's id and body for printing.
func (h Hit) Title() (any, string) {
	id := h.ID
	if id == nil {
		id = h.PK
	}
	body, _ := h.Source["body"].(string)
	return id, body
}
