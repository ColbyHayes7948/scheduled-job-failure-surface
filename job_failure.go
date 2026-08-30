package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"scheduled-job-failures/infrai"
)

func main() {
	client, err := infrai.NewFromEnvironment()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = runScheduledJob()
	if err == nil {
		fmt.Println("shipment-sync completed")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if captureErr := client.Capture(ctx, infrai.ErrorPayload{
		Message:     err.Error(),
		Level:       "error",
		Fingerprint: []string{"shipment-sync", "scheduled-job"},
		Exception:   err.Error(),
		Context: map[string]any{
			"job":         "shipment-sync",
			"schedule":    "every 15 minutes",
			"occurred_at": time.Now().UTC().Format(time.RFC3339),
		},
	}); captureErr != nil {
		fmt.Fprintln(os.Stderr, captureErr)
		os.Exit(1)
	}
	fmt.Println("shipment-sync failure captured")
}

func runScheduledJob() error {
	return fmt.Errorf("warehouse API returned an empty shipment batch")
}
