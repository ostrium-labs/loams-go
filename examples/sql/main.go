// Command sql runs read-only SQL over a collection, including the rrf,
// vector_search and text_search table functions.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ostrium-labs/loams-go/internal/rest"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "sql:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	api := rest.New()
	if err := api.Setup(ctx, "docs"); err != nil {
		return err
	}
	queries := []string{
		"SELECT tenant, count(*) AS n FROM docs GROUP BY tenant ORDER BY tenant",
		"SELECT n, body FROM docs WHERE tenant = 'b' ORDER BY n",
	}
	if q := os.Getenv("LOAMS_SQL"); q != "" {
		queries = []string{q}
	}
	for _, q := range queries {
		fmt.Println(">", q)
		cols, rows, err := api.SQL(ctx, q)
		if err != nil {
			return err
		}
		fmt.Println(strings.Join(cols, "\t"))
		for _, r := range rows {
			cells := make([]string, len(r))
			for i, c := range r {
				cells[i] = fmt.Sprint(c)
			}
			fmt.Println(strings.Join(cells, "\t"))
		}
	}
	return nil
}
