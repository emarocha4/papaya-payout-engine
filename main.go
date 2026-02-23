package main

import (
	"log"

	"github.com/yuno-payments/papaya-payout-engine/cmd/server"
)

func main() {
	srv, err := server.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
