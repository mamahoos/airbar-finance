// Dev/integration helper: calls FinanceHealthService.CheckReady.
// Usage: GRPC_ADDR=localhost:50051 go run ./scripts/check_ready
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	financev1 "github.com/mamahoos/airbar-finance/internal/gen/financev1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := os.Getenv("GRPC_ADDR")
	if addr == "" {
		addr = "localhost:50051"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "grpc client: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := financev1.NewFinanceHealthServiceClient(conn)
	resp, err := client.CheckReady(ctx, &financev1.HealthCheckRequest{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "CheckReady: %v\n", err)
		os.Exit(1)
	}
	if !resp.Ready {
		fmt.Fprintf(os.Stderr, "service not ready\n")
		os.Exit(1)
	}

	fmt.Println("ready=true")
}
