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
		if text == protocol.CommandQuit {
			return nil
		}
		if text == "" {
			continue
		}
		message, ok := parseInput(text)
		if !ok {
			message = protocol.Message{Type: protocol.TypeChat, Content: text}
		}
		if err := encoder.Encode(message); err != nil {
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

func parseInput(text string) (protocol.Message, bool) {
	parts := strings.Fields(text)
	if len(parts) == 0 || !strings.HasPrefix(parts[0], "/") {
		return protocol.Message{}, false
	}
	switch parts[0] {
	case protocol.CommandRooms:
		return protocol.Message{Type: protocol.TypeListRooms}, true
	case protocol.CommandJoin:
		return protocol.Message{Type: protocol.TypeJoinRoom, Content: strings.TrimSpace(strings.TrimPrefix(text, parts[0]))}, true
	case protocol.CommandCreate:
		return protocol.Message{Type: protocol.TypeCreateRoom, Content: strings.TrimSpace(strings.TrimPrefix(text, parts[0]))}, true
	case protocol.CommandLeave:
		return protocol.Message{Type: protocol.TypeLeaveRoom}, true
	default:
		return protocol.Message{Type: protocol.TypeChat, Content: text}, false
	}
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
