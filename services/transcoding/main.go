package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Create a context that will be canceled on SIGTERM/SIGINT
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize infrastructure
	dbPool := infrastructure.NewPool()
	defer dbPool.Close()

	db := infrastructure.NewDatabase(dbPool)

	consumer, err := infrastructure.NewConsumer()
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	publisher, err := infrastructure.NewPublisher()
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	// Start consumer in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting down consumer...")
				return
			default:
				if err := consumeMessage(ctx, consumer, db, publisher); err != nil {
					log.Printf("Error consuming message: %v", err)
					// Add delay to prevent CPU spinning on repeated errors
					time.Sleep(time.Second)
				}
			}
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Received shutdown signal")
	cancel() // This will trigger graceful shutdown

	// Give some time for ongoing operations to complete
	time.Sleep(time.Second * 5)
	log.Println("Service stopped")
}
