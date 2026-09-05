package store

import "github.com/snna2069/TerminalChatApplication/pkg/models"

type MessageStore interface {
	SaveMessage(message models.Message) error
	GetRoomHistory(room string, limit int) ([]models.Message, error)
}
