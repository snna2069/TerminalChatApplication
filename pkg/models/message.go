package models

import "time"

type Message struct {
	Timestamp time.Time `json:"timestamp"`
	Username  string    `json:"username"`
	Room      string    `json:"room"`
	Content   string    `json:"content"`
}
