// Command platform lists Live apps using Go's 0.1.2 LiveAdmin module.
// Auth and general admin modules are not exposed by the Go SDK in 0.1.2.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	loams "loams.dev/go"
	livev1 "loams.dev/go/gen/loams/live/v1"
)

func main() {
	if os.Getenv("LOAMS_PLATFORM_EXAMPLE") != "1" {
		log.Fatal("set LOAMS_PLATFORM_EXAMPLE=1 to run this read-only platform example")
	}
	if os.Getenv("LOAMS_API_KEY") == "" && os.Getenv("LOAMS_TOKEN") == "" {
		log.Fatal("set LOAMS_API_KEY or LOAMS_TOKEN to an authorized credential")
	}
	if err := run(context.Background()); err != nil {
		log.Fatal("platform: ", err)
	}
}

func run(ctx context.Context) error {
	client, err := loams.FromEnv(loams.Options{})
	if err != nil {
		return err
	}
	defer client.Close()
	available, err := client.System().Available(ctx, "live")
	if err != nil {
		return err
	}
	if !available {
		fmt.Println("Live admin is unavailable in this engine build")
		return nil
	}
	apps, err := client.LiveAdmin().ListApps(ctx, &livev1.ListAppsRequest{PageSize: 10})
	if err != nil {
		return err
	}
	for _, app := range apps.GetApps() {
		fmt.Println(app.GetName())
	}
	return nil
}
