package main

import (
	"flag"
	"log"

	"github.com/snna2069/TerminalChatApplication/internal/client"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "TCP server address")
	flag.Parse()

	if err := client.Run(*addr); err != nil {
		log.Println(err)
	}
}
