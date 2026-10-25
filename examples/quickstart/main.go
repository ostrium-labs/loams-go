// Command quickstart connects with the Go SDK, creates a collection, upserts
// documents and runs a hybrid search (vector + BM25 + filter).
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	loams "loams.dev/go"

	"github.com/ostrium-labs/loams-go/internal/rest"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "quickstart:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// The SDK: ask the instance what it is.
	client, err := loams.New(loams.Options{Endpoint: rest.Endpoint(), Auth: loams.APIKey(os.Getenv("LOAMS_API_KEY"))})
	if err != nil {
		return err
	}
	defer client.Close()
	info, err := client.Instance().GetInstance(ctx, &loams.GetInstanceRequest{})
	if err != nil {
		return fmt.Errorf("is an engine running at %s? %w", rest.Endpoint(), err)
	}
	fmt.Printf("connected to %s %s (%s)\n", info.GetName(), info.GetServerVersion(), info.GetEdition())

	// Demonstrate collections and search through the native REST API.
	api := rest.New()
	if err := api.Setup(ctx, "docs"); err != nil {
		return err
	}
	fmt.Printf("wrote 4 documents to %s/docs\n", rest.Namespace)

	// Query vector near doc 1, text "refund", only tenant "a" documents.
	res, err := api.HybridSearch(ctx, "docs", [3]float64{1, 0, 0}, "refund", "a", 3)
	if err != nil {
		return err
	}
	for i, hit := range res.Hits {
		id, body := hit.Title()
		fmt.Printf("%d. id=%v score=%.4f %s\n", i+1, id, hit.Score, body)
	}
	return nil
}
