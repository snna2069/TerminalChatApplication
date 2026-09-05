package server

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/snna2069/TerminalChatApplication/internal/protocol"
)

func TestTCPChatIntegration(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	chatServer := New(listener.Addr().String())
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- chatServer.Serve(listener)
	}()

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := chatServer.Shutdown(ctx); err != nil {
			t.Error(err)
		}
		if err := <-serverErrors; err != nil {
			t.Errorf("server returned %v", err)
		}
	}()

	alice := connectTestClient(t, listener.Addr().String(), "alice")
	defer alice.conn.Close()
	bob := connectTestClient(t, listener.Addr().String(), "bob")
	defer bob.conn.Close()
	readTestMessage(t, alice)

	writeTestMessage(t, alice, protocol.Message{Type: protocol.TypeChat, Content: "hello integration"})
	received := readTestMessage(t, bob)
	if received.Type != protocol.TypeChat || received.Username != "alice" || received.Content != "hello integration" {
		t.Fatalf("received %#v", received)
	}
}

type testClient struct {
	conn    net.Conn
	encoder *json.Encoder
	decoder *json.Decoder
}

func connectTestClient(t *testing.T, address, username string) *testClient {
	t.Helper()
	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	client := &testClient{conn: conn, encoder: json.NewEncoder(conn), decoder: json.NewDecoder(conn)}
	writeTestMessage(t, client, protocol.Message{Type: protocol.TypeRegister, Username: username})
	message := readTestMessage(t, client)
	if message.Type != protocol.TypeSystem {
		t.Fatalf("registration response = %#v", message)
	}
	return client
}

func writeTestMessage(t *testing.T, client *testClient, message protocol.Message) {
	t.Helper()
	if err := client.encoder.Encode(message); err != nil {
		t.Fatal(err)
	}
}

func readTestMessage(t *testing.T, client *testClient) protocol.Message {
	t.Helper()
	if err := client.conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	var message protocol.Message
	if err := client.decoder.Decode(&message); err != nil {
		t.Fatal(err)
	}
	return message
}
