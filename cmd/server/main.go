package main

import (
	"flag"
	"log"

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
	if err := server.NewWithStore(*addr, messageStore).ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
