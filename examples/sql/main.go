// Command sql runs read-only SQL over the quickstart collection through loams.sql.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	loams "loams.dev/go"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal("sql: ", err)
	}
}

func run(ctx context.Context) error {
	namespace := os.Getenv("LOAMS_NAMESPACE")
	if namespace == "" {
		namespace = "examples-go"
	}
	client, err := loams.FromEnv(loams.Options{Namespace: namespace})
	if err != nil {
		return err
	}
	defer client.Close()
	queries := []struct {
		statement string
		params    []any
	}{
		{"SELECT tenant, count(*) AS n FROM kb GROUP BY tenant ORDER BY tenant", nil},
		{"SELECT n, body FROM kb WHERE tenant = $1 ORDER BY n", []any{"a"}},
		{"SELECT _id, body, _score FROM rrf(vector_search('kb', [1.0, 0.0, 0.0], 'embedding', 10), text_search('kb', 'refund', 'body', 10)) LIMIT 3", nil},
	}
	if query := os.Getenv("LOAMS_SQL"); query != "" {
		queries = []struct {
			statement string
			params    []any
		}{{statement: query}}
	}
	for _, query := range queries {
		fmt.Println(">", query.statement)
		rows, err := client.Sql().Records(ctx, query.statement, query.params)
		if err != nil {
			return err
		}
		for _, row := range rows {
			fmt.Println(row)
		}
	}
	return nil
}
