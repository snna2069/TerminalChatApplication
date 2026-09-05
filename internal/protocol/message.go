package protocol

import "encoding/json"

const (
	TypeRegister   = "register"
	TypeChat       = "chat"
	TypeCreateRoom = "create_room"
	TypeJoinRoom   = "join_room"
	TypeLeaveRoom  = "leave_room"
	TypeListRooms  = "list_rooms"
	TypeListUsers  = "list_users"
	TypePrivate    = "private_message"
	TypeSystem     = "system"
	TypeError      = "error"
)

type Message struct {
	Type     string `json:"type"`
	Username string `json:"username,omitempty"`
	Target   string `json:"target,omitempty"`
	Room     string `json:"room,omitempty"`
	Content  string `json:"content,omitempty"`
}

func Encode(message Message) ([]byte, error) {
	return json.Marshal(message)
}
