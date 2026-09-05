package protocol

import "encoding/json"

const (
	TypeRegister = "register"
	TypeChat     = "chat"
	TypeSystem   = "system"
	TypeError    = "error"
)

type Message struct {
	Type     string `json:"type"`
	Username string `json:"username,omitempty"`
	Room     string `json:"room,omitempty"`
	Content  string `json:"content,omitempty"`
}

func Encode(message Message) ([]byte, error) {
	return json.Marshal(message)
}
