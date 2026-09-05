package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/snna2069/TerminalChatApplication/pkg/models"
)

func TestJSONLStoreRoundTripAndLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "messages.jsonl")
	messageStore, err := NewJSONLStore(path)
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	messages := []models.Message{
		{Timestamp: base, Username: "alice", Room: "general", Content: "one"},
		{Timestamp: base.Add(time.Minute), Username: "bob", Room: "lounge", Content: "other room"},
		{Timestamp: base.Add(2 * time.Minute), Username: "alice", Room: "general", Content: "two"},
	}
	for _, message := range messages {
		if err := messageStore.SaveMessage(message); err != nil {
			t.Fatal(err)
		}
	}

	history, err := messageStore.GetRoomHistory("general", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].Content != "two" {
		t.Fatalf("history = %#v, want latest general message", history)
	}
}

func TestMemoryStoreReturnsRoomHistory(t *testing.T) {
	messageStore := NewMemoryStore()
	for index, content := range []string{"one", "two", "three"} {
		if err := messageStore.SaveMessage(models.Message{
			Timestamp: time.Unix(int64(index), 0),
			Room:      "general",
			Content:   content,
		}); err != nil {
			t.Fatal(err)
		}
	}
	history, err := messageStore.GetRoomHistory("general", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 2 || history[0].Content != "two" || history[1].Content != "three" {
		t.Fatalf("history = %#v, want two latest messages", history)
	}
}
