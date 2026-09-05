package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/snna2069/TerminalChatApplication/internal/server"
	"github.com/snna2069/TerminalChatApplication/internal/store"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "TCP address to listen on")
	dataPath := flag.String("data", "data/messages.jsonl", "JSON Lines message history path")
	flag.Parse()

	messageStore, err := store.NewJSONLStore(*dataPath)
	if err != nil {
		log.Fatal(err)
	}
	chatServer := server.NewWithStore(*addr, messageStore)
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- chatServer.ListenAndServe()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	select {
	case err := <-serverErrors:
		if err != nil {
			log.Fatal(err)
		}
	case signal := <-signals:
		log.Printf("received %s, shutting down", signal)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := chatServer.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}
}
