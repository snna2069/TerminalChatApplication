package store

import (
	"sort"
	"sync"

	"github.com/snna2069/TerminalChatApplication/pkg/models"
)

type MemoryStore struct {
	mu       sync.RWMutex
	messages []models.Message
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) SaveMessage(message models.Message) error {
	s.mu.Lock()
	s.messages = append(s.messages, message)
	s.mu.Unlock()
	return nil
}

func (s *MemoryStore) GetRoomHistory(room string, limit int) ([]models.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 {
		return []models.Message{}, nil
	}
	result := make([]models.Message, 0, limit)
	for index := len(s.messages) - 1; index >= 0 && len(result) < limit; index-- {
		if s.messages[index].Room == room {
			result = append(result, s.messages[index])
		}
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].Timestamp.Before(result[right].Timestamp)
	})
	return result, nil
}
