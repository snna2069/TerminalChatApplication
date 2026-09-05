package server

import (
	"strings"
	"testing"
	"time"

	"github.com/snna2069/TerminalChatApplication/internal/protocol"
	"github.com/snna2069/TerminalChatApplication/pkg/models"
)

func TestRoomMembership(t *testing.T) {
	s := New("localhost:0")
	first := &client{out: make(chan protocol.Message, 8)}
	second := &client{out: make(chan protocol.Message, 8)}
	if err := s.registerClient(first, "alice"); err != nil {
		t.Fatal(err)
	}
	if err := s.registerClient(second, "bob"); err != nil {
		t.Fatal(err)
	}

	s.createRoom(first, "lounge")
	if first.room != "lounge" {
		t.Fatalf("first client room = %q, want lounge", first.room)
	}
	duplicate := &client{out: make(chan protocol.Message, 1)}
	if err := s.registerClient(duplicate, "alice"); err == nil {
		t.Fatal("expected duplicate username error")
	}

	s.joinRoom(second, "lounge")
	if len(s.rooms["lounge"].clients) != 2 {
		t.Fatalf("lounge members = %d, want 2", len(s.rooms["lounge"].clients))
	}
	s.leaveRoom(second)
	if second.room != protocol.DefaultRoom {
		t.Fatalf("second client room = %q, want %s", second.room, protocol.DefaultRoom)
	}
	if len(s.rooms["lounge"].clients) != 1 {
		t.Fatalf("lounge members after leave = %d, want 1", len(s.rooms["lounge"].clients))
	}
}

func TestBroadcastToRoomOnlyReachesMembers(t *testing.T) {
	s := New("localhost:0")
	first := &client{out: make(chan protocol.Message, 8)}
	second := &client{out: make(chan protocol.Message, 8)}
	if err := s.registerClient(first, "alice"); err != nil {
		t.Fatal(err)
	}
	if err := s.registerClient(second, "bob"); err != nil {
		t.Fatal(err)
	}
	s.createRoom(first, "lounge")
	for len(first.out) > 0 {
		<-first.out
	}

	s.broadcastToRoom("lounge", protocol.Message{Type: protocol.TypeChat, Content: "private room"})
	select {
	case message := <-first.out:
		if message.Content != "private room" {
			t.Fatalf("first received %q, want private room", message.Content)
		}
	default:
		t.Fatal("room member did not receive message")
	}
	select {
	case message := <-second.out:
		t.Fatalf("general room member received %#v", message)
	default:
	}
}

func TestPrivateMessage(t *testing.T) {
	s := New("localhost:0")
	alice := &client{out: make(chan protocol.Message, 4)}
	bob := &client{out: make(chan protocol.Message, 4)}
	if err := s.registerClient(alice, "alice"); err != nil {
		t.Fatal(err)
	}
	if err := s.registerClient(bob, "bob"); err != nil {
		t.Fatal(err)
	}

	s.privateMessage(alice, "bob", "hello privately")
	received := <-bob.out
	if received.Type != protocol.TypePrivate || received.Content != "hello privately" {
		t.Fatalf("bob received %#v", received)
	}
	confirmed := <-alice.out
	if confirmed.Target != "bob" || confirmed.Content != "hello privately" {
		t.Fatalf("alice received confirmation %#v", confirmed)
	}

	s.privateMessage(alice, "nobody", "hello")
	errorMessage := <-alice.out
	if errorMessage.Type != protocol.TypeError {
		t.Fatalf("missing-user response = %#v", errorMessage)
	}
}

func TestHistoryUsesCurrentRoom(t *testing.T) {
	s := New("localhost:0")
	alice := &client{out: make(chan protocol.Message, 8)}
	if err := s.registerClient(alice, "alice"); err != nil {
		t.Fatal(err)
	}
	if err := s.store.SaveMessage(models.Message{
		Timestamp: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC),
		Username:  "alice",
		Room:      protocol.DefaultRoom,
		Content:   "persisted hello",
	}); err != nil {
		t.Fatal(err)
	}

	s.history(alice, 10)
	response := <-alice.out
	if response.Type != protocol.TypeSystem || !strings.Contains(response.Content, "persisted hello") {
		t.Fatalf("history response = %#v", response)
	}
}
