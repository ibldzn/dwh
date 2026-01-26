package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/ibldzn/dwh-v2/internal/fetcher"
	"github.com/ibldzn/dwh-v2/internal/fincloud"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		errorExit("failed to load .env file: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	client, err := fincloud.NewClient(fincloud.Config{})
	if err != nil {
		errorExit("failed to create fincloud client")
	}

	session, err := client.Login(ctx, fincloud.Credentials{
		Username:   requireEnv("FINCLOUD_USERNAME"),
		Password:   requireEnv("FINCLOUD_PASSWORD"),
		RoleID:     requireEnv("FINCLOUD_ROLE_ID"),
		LocationID: requireEnv("FINCLOUD_LOCATION_ID"),
	})
	if err != nil {
		errorExit("failed to login", err)
	}

	ctx = fincloud.WithFincloudSessionID(ctx, session.ID)

	fetch, err := fetcher.NewFetcher(client)
	if err != nil {
		errorExit("failed to create fetcher", err)
	}

	cifs, err := fetch.FetchCIFList(ctx)
	if err != nil {
		errorExit("failed to fetch CIF detail", err)
	}

	fmt.Printf("fetched %d CIFs\n", len(cifs))

	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup

cifLoop:
	for _, cifNo := range cifs {
		select {
		case <-ctx.Done():
			break cifLoop
		case sem <- struct{}{}:
			wg.Go(func() {
				defer func() { <-sem }()

				cif, err := fetch.FetchCIFDetail(ctx, cifNo)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to fetch CIF %s: %v\n", cifNo, err)
					return
				}

				fmt.Printf("CIF %s: %+v\n", cifNo, cif.NamaNasabah)
			})
		}
	}

	wg.Wait()

	fmt.Println("done")
}

func requireEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		errorExit("environment variable %q is required", key)
	}
	return val
}

func errorExit(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
