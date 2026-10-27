// Command quickstart connects with the Go SDK, creates an examples namespace
// and kb collection, then exercises collections, documents, search, vector,
// SQL, streams and optional Graph through their SDK modules.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	loams "loams.dev/go"
	collectionv1 "loams.dev/go/gen/loams/collection/v1"
	graphv1 "loams.dev/go/gen/loams/graph/v1"
	streamv1 "loams.dev/go/gen/loams/stream/v1"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal("quickstart: ", err)
	}
}

func run(ctx context.Context) error {
	namespace := os.Getenv("LOAMS_NAMESPACE")
	if namespace == "" {
		namespace = "examples-go"
	}
	client, err := loams.FromEnv(loams.Options{Namespace: namespace, SessionConsistency: true})
	if err != nil {
		return err
	}
	defer client.Close()

	info, err := client.Instance().GetInstance(ctx, &loams.GetInstanceRequest{})
	if err != nil {
		return fmt.Errorf("is an engine running? %w", err)
	}
	fmt.Printf("connected to %s %s (namespace %s)\n", info.GetName(), info.GetServerVersion(), namespace)

	if _, err := client.Namespaces().Create(ctx, &collectionv1.CreateNamespaceRequest{Namespace: namespace}); err != nil && !alreadyExists(err) {
		return err
	}
	schema, err := loams.NewSchema().Text("body").Keyword("tenant").I64("n").Vector("embedding", 3).Struct()
	if err != nil {
		return err
	}
	if _, err := client.Collections().Create(ctx, &collectionv1.CreateCollectionRequest{
		Namespace: namespace, Name: "kb", Schema: schema,
	}); err != nil && !alreadyExists(err) {
		return err
	}

	kb := client.Collection("kb", namespace)
	if _, err := kb.Upsert(ctx, []loams.DocumentInput{
		{ID: loams.UintID(1), Source: loams.JSONObject{"body": "refund policy", "tenant": "a", "n": 1}, Vectors: map[string][]float32{"embedding": {1, 0, 0}}},
		{ID: loams.UintID(2), Source: loams.JSONObject{"body": "shipping times", "tenant": "a", "n": 2}, Vectors: map[string][]float32{"embedding": {0.9, 0.1, 0}}},
		{ID: loams.UintID(3), Source: loams.JSONObject{"body": "refund window", "tenant": "b", "n": 3}, Vectors: map[string][]float32{"embedding": {0, 0, 1}}},
	}); err != nil {
		return err
	}

	hits, err := kb.Search(ctx, kb.Query().
		Text("refund", loams.TextOptions{Fields: []string{"body"}}).
		Vector("embedding", []float32{1, 0, 0}, loams.VectorOptions{K: 10}).
		Fuse(loams.RRF(60)).
		Filter(loams.Filter.Eq("tenant", "a")).
		Limit(3))
	if err != nil {
		return err
	}
	for _, hit := range hits.GetHits() {
		fmt.Println("hybrid", hit.GetPk(), hit.GetScore(), hit.GetSource().AsMap())
	}

	// Use the generated loams.vector module with a builder-produced request.
	vectorRequest, err := loams.NewSearch("kb").
		Vector("embedding", []float32{1, 0, 0}, loams.VectorOptions{K: 3}).
		Filter(loams.Filter.Eq("tenant", "a")).Limit(3).Build()
	if err != nil {
		return err
	}
	vector, err := client.Vector().Search(ctx, vectorRequest)
	if err != nil {
		return err
	}
	fmt.Println("vector hits:", len(vector.GetHits()))

	rows, err := client.Sql().Records(ctx,
		"SELECT tenant, count(*) AS n FROM kb WHERE n >= $1 GROUP BY tenant ORDER BY tenant", []any{int64(1)})
	if err != nil {
		return err
	}
	for _, row := range rows {
		fmt.Println("sql", row["tenant"], row["n"])
	}

	if _, err := client.Streams().Create(ctx, &streamv1.CreateStreamRequest{Name: "kb-events", Partitions: 1}); err != nil {
		return err
	}
	producer, sequence := "go-example", uint64(1)
	if _, err := client.Streams().Produce(ctx, &streamv1.ProduceRequest{
		Stream: "kb-events", Records: []*streamv1.Record{{Value: []byte(`{"document":1,"action":"indexed"}`)}},
		ProducerId: &producer, ProducerSeq: &sequence,
	}); err != nil {
		return err
	}
	events := client.Streams().Subscribe(ctx, &streamv1.FetchRequest{Stream: "kb-events"})
	defer events.Close()
	if events.Next() {
		fmt.Println("stream event offset:", events.Record().GetOffset())
	}
	if err := events.Err(); err != nil {
		return err
	}

	// Graph is experimental and feature-gated. It is queried only when this
	// engine advertises the module and this namespace already has a graph.
	graphEnabled, err := client.System().Available(ctx, "graph")
	if err != nil {
		return err
	}
	if graphEnabled {
		graphs, err := client.Graph().List(ctx, &graphv1.ListGraphsRequest{Namespace: namespace, PageSize: 1})
		if err != nil {
			return err
		}
		if len(graphs.GetGraphs()) > 0 {
			result, err := client.Graph().Query(ctx, &graphv1.ExecuteRequest{
				Namespace: namespace, Graph: graphs.GetGraphs()[0].GetName(),
				Statement: "RETURN 1 AS value", Language: graphv1.QueryLanguage_QUERY_LANGUAGE_GQL, ReadOnly: true,
			})
			if err != nil {
				return err
			}
			fmt.Println("graph rows:", len(result.GetRows().GetRows()))
		} else {
			fmt.Println("graph is enabled; create a graph before running the optional GQL query")
		}
	} else {
		fmt.Println("graph is experimental and not enabled in this engine build")
	}
	return nil
}

func alreadyExists(err error) bool {
	var exists *loams.AlreadyExistsError
	return errors.As(err, &exists)
}
