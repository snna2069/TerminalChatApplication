package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/snna2069/TerminalChatApplication/pkg/models"
)

type JSONLStore struct {
	mu   sync.Mutex
	path string
}

func NewJSONLStore(path string) (*JSONLStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create message data directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("create message data file: %w", err)
	}
	file.Close()
	return &JSONLStore{path: path}, nil
}

func (s *JSONLStore) SaveMessage(message models.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.OpenFile(s.path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open message data file: %w", err)
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(message); err != nil {
		return fmt.Errorf("write message data: %w", err)
	}
	return nil
}

func (s *JSONLStore) GetRoomHistory(room string, limit int) ([]models.Message, error) {
	if limit <= 0 {
		return []models.Message{}, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.Open(s.path)
	if err != nil {
		return nil, fmt.Errorf("open message data file: %w", err)
	}
	defer file.Close()

	all := make([]models.Message, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var message models.Message
		if err := json.Unmarshal(scanner.Bytes(), &message); err != nil {
			return nil, fmt.Errorf("decode message data: %w", err)
		}
		if message.Room == room {
			all = append(all, message)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read message data: %w", err)
	}
	if len(all) > limit {
		all = all[len(all)-limit:]
	}
	return all, nil
}
