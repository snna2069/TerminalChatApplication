package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/snna2069/TerminalChatApplication/internal/protocol"
)

func Run(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("connect to %s: %w", addr, err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(bufio.NewReader(conn))
	input := bufio.NewScanner(os.Stdin)

	fmt.Print("Welcome to Terminal Chat!\n\nEnter username: ")
	if !input.Scan() {
		return input.Err()
	}
	username := strings.TrimSpace(input.Text())
	if err := encoder.Encode(protocol.Message{Type: protocol.TypeRegister, Username: username}); err != nil {
		return err
	}

	readDone := make(chan error, 1)
	go func() {
		for {
			var message protocol.Message
			if err := decoder.Decode(&message); err != nil {
				if err == io.EOF {
					readDone <- fmt.Errorf("server disconnected")
				} else {
					readDone <- err
				}
				return
			}
			display(message)
		}
	}()

	fmt.Println("Type a message and press Enter. Use /quit to exit.")
	for input.Scan() {
		text := strings.TrimSpace(input.Text())
		if text == "/quit" {
			return nil
		}
		if text == "" {
			continue
		}
		if err := encoder.Encode(protocol.Message{Type: protocol.TypeChat, Content: text}); err != nil {
			return err
		}
		select {
		case err := <-readDone:
			return err
		default:
		}
	}
	if err := input.Err(); err != nil {
		return err
	}
	return <-readDone
}

func display(message protocol.Message) {
	switch message.Type {
	case protocol.TypeChat:
		fmt.Printf("[%s@%s] %s\n> ", message.Username, message.Room, message.Content)
	case protocol.TypeError:
		fmt.Printf("[error] %s\n> ", message.Content)
	default:
		fmt.Printf("[system] %s\n> ", message.Content)
	}
}
