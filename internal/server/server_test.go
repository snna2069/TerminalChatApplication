package server

import (
	"testing"

	"github.com/snna2069/TerminalChatApplication/internal/protocol"
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
