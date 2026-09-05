package main

import (
	"flag"
	"log"

	"github.com/snna2069/TerminalChatApplication/internal/server"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "TCP address to listen on")
	flag.Parse()

	if err := server.New(*addr).ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
