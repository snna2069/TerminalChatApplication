package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/snna2069/TerminalChatApplication/internal/protocol"
	"github.com/snna2069/TerminalChatApplication/internal/store"
	"github.com/snna2069/TerminalChatApplication/pkg/models"
)

type Server struct {
	addr    string
	mu      sync.RWMutex
	clients map[*client]struct{}
	rooms   map[string]*room
	store   store.MessageStore
}

type client struct {
	conn     net.Conn
	username string
	room     string
	out      chan protocol.Message
}

func New(addr string) *Server {
	return NewWithStore(addr, store.NewMemoryStore())
}

func NewWithStore(addr string, messageStore store.MessageStore) *Server {
	return &Server{
		addr:    addr,
		clients: make(map[*client]struct{}),
		rooms:   map[string]*room{protocol.DefaultRoom: newRoom(protocol.DefaultRoom)},
		store:   messageStore,
	}
}

func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("chat server listening on %s", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	c := &client{conn: conn, out: make(chan protocol.Message, 16)}
	defer conn.Close()

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		encoder := json.NewEncoder(conn)
		for message := range c.out {
			if err := encoder.Encode(message); err != nil {
				return
			}
		}
	}()

	decoder := json.NewDecoder(bufio.NewReader(conn))
	registered := false
	defer func() {
		if registered {
			s.removeClient(c)
			s.broadcast(protocol.Message{Type: protocol.TypeSystem, Content: fmt.Sprintf("%s disconnected", c.username)}, c)
		}
		close(c.out)
		<-writerDone
	}()

	for {
		var message protocol.Message
		if err := decoder.Decode(&message); err != nil {
			if err != io.EOF {
				log.Printf("connection from %s ended: %v", conn.RemoteAddr(), err)
			}
			return
		}

		if !registered {
			if message.Type != protocol.TypeRegister {
				c.send(protocol.Message{Type: protocol.TypeError, Content: "first message must register a username"})
				continue
			}
			if err := s.registerClient(c, message.Username); err != nil {
				c.send(protocol.Message{Type: protocol.TypeError, Content: err.Error()})
				continue
			}
			registered = true
			c.send(protocol.Message{Type: protocol.TypeSystem, Room: c.room, Content: fmt.Sprintf("connected as %s; joined room %s", c.username, c.room)})
			s.broadcast(protocol.Message{Type: protocol.TypeSystem, Content: fmt.Sprintf("%s joined the chat", c.username)}, c)
			continue
		}

		switch message.Type {
		case protocol.TypeChat:
			content := strings.TrimSpace(message.Content)
			if content == "" {
				c.send(protocol.Message{Type: protocol.TypeError, Content: "message cannot be empty"})
				continue
			}
			if err := s.store.SaveMessage(models.Message{
				Timestamp: time.Now().UTC(),
				Username:  c.username,
				Room:      c.room,
				Content:   content,
			}); err != nil {
				log.Printf("persist message from %s: %v", c.username, err)
				c.send(protocol.Message{Type: protocol.TypeError, Content: "message could not be saved"})
				continue
			}
			s.broadcastToRoom(c.room, protocol.Message{Type: protocol.TypeChat, Username: c.username, Room: c.room, Content: content})
		case protocol.TypeCreateRoom:
			s.createRoom(c, message.Content)
		case protocol.TypeJoinRoom:
			s.joinRoom(c, message.Content)
		case protocol.TypeLeaveRoom:
			s.leaveRoom(c)
		case protocol.TypeListRooms:
			s.listRooms(c)
		case protocol.TypeListUsers:
			s.listUsers(c)
		case protocol.TypePrivate:
			s.privateMessage(c, message.Target, message.Content)
		case protocol.TypeHistory:
			s.history(c, message.Limit)
		default:
			c.send(protocol.Message{Type: protocol.TypeError, Content: "unsupported message type"})
		}
	}
}

func (s *Server) registerClient(c *client, username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if strings.ContainsAny(username, " \t\r\n") {
		return fmt.Errorf("username cannot contain whitespace")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for existing := range s.clients {
		if strings.EqualFold(existing.username, username) {
			return fmt.Errorf("username %q is already in use", username)
		}
	}
	c.username = username
	c.room = protocol.DefaultRoom
	s.clients[c] = struct{}{}
	s.rooms[c.room].clients[c] = struct{}{}
	return nil
}

func (s *Server) removeClient(c *client) {
	s.mu.Lock()
	delete(s.clients, c)
	if current, ok := s.rooms[c.room]; ok {
		delete(current.clients, c)
	}
	s.mu.Unlock()
}

func (s *Server) broadcast(message protocol.Message, excluded *client) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for c := range s.clients {
		if c != excluded {
			c.send(message)
		}
	}
}

func (s *Server) broadcastToRoom(roomName string, message protocol.Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	current, ok := s.rooms[roomName]
	if !ok {
		return
	}
	for c := range current.clients {
		c.send(message)
	}
}

func (s *Server) createRoom(c *client, name string) {
	name = strings.TrimSpace(name)
	if err := validateRoomName(name); err != nil {
		c.send(protocol.Message{Type: protocol.TypeError, Content: err.Error()})
		return
	}

	s.mu.Lock()
	if _, exists := s.rooms[name]; exists {
		s.mu.Unlock()
		c.send(protocol.Message{Type: protocol.TypeError, Content: fmt.Sprintf("room %q already exists", name)})
		return
	}
	s.rooms[name] = newRoom(name)
	s.mu.Unlock()
	c.send(protocol.Message{Type: protocol.TypeSystem, Content: fmt.Sprintf("created room %s", name)})
	s.joinRoom(c, name)
}

func (s *Server) joinRoom(c *client, name string) {
	name = strings.TrimSpace(name)
	if err := validateRoomName(name); err != nil {
		c.send(protocol.Message{Type: protocol.TypeError, Content: err.Error()})
		return
	}

	s.mu.Lock()
	target, exists := s.rooms[name]
	if !exists {
		s.mu.Unlock()
		c.send(protocol.Message{Type: protocol.TypeError, Content: fmt.Sprintf("room %q does not exist", name)})
		return
	}
	if c.room == name {
		s.mu.Unlock()
		c.send(protocol.Message{Type: protocol.TypeSystem, Room: name, Content: fmt.Sprintf("already in room %s", name)})
		return
	}
	previous := c.room
	if current, ok := s.rooms[previous]; ok {
		delete(current.clients, c)
	}
	target.clients[c] = struct{}{}
	c.room = name
	s.mu.Unlock()

	c.send(protocol.Message{Type: protocol.TypeSystem, Room: name, Content: fmt.Sprintf("joined room %s", name)})
	s.broadcastToRoom(name, protocol.Message{Type: protocol.TypeSystem, Room: name, Content: fmt.Sprintf("%s joined the room", c.username)})
}

func (s *Server) leaveRoom(c *client) {
	s.mu.RLock()
	current := c.room
	s.mu.RUnlock()
	if current == protocol.DefaultRoom {
		c.send(protocol.Message{Type: protocol.TypeError, Room: current, Content: "you are already in the general room"})
		return
	}
	s.joinRoom(c, protocol.DefaultRoom)
}

func (s *Server) listRooms(c *client) {
	s.mu.RLock()
	names := make([]string, 0, len(s.rooms))
	for name := range s.rooms {
		names = append(names, name)
	}
	s.mu.RUnlock()
	sort.Strings(names)
	c.send(protocol.Message{Type: protocol.TypeSystem, Content: "rooms: " + strings.Join(names, ", ")})
}

func (s *Server) listUsers(c *client) {
	s.mu.RLock()
	users := make([]string, 0, len(s.clients))
	for current := range s.clients {
		users = append(users, fmt.Sprintf("%s (%s)", current.username, current.room))
	}
	s.mu.RUnlock()
	sort.Strings(users)
	c.send(protocol.Message{Type: protocol.TypeSystem, Content: "online users:\n- " + strings.Join(users, "\n- ")})
}

func (s *Server) privateMessage(sender *client, targetName, content string) {
	targetName = strings.TrimSpace(targetName)
	content = strings.TrimSpace(content)
	if targetName == "" {
		sender.send(protocol.Message{Type: protocol.TypeError, Content: "usage: /msg username message"})
		return
	}
	if content == "" {
		sender.send(protocol.Message{Type: protocol.TypeError, Content: "private message cannot be empty"})
		return
	}

	s.mu.RLock()
	var recipient *client
	for current := range s.clients {
		if strings.EqualFold(current.username, targetName) {
			recipient = current
			break
		}
	}
	s.mu.RUnlock()
	if recipient == nil {
		sender.send(protocol.Message{Type: protocol.TypeError, Content: fmt.Sprintf("user %q is offline", targetName)})
		return
	}
	private := protocol.Message{Type: protocol.TypePrivate, Username: sender.username, Target: recipient.username, Content: content}
	recipient.send(private)
	sender.send(private)
}

func (s *Server) history(c *client, limit int) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	messages, err := s.store.GetRoomHistory(c.room, limit)
	if err != nil {
		log.Printf("load history for %s: %v", c.room, err)
		c.send(protocol.Message{Type: protocol.TypeError, Content: "history could not be loaded"})
		return
	}
	if len(messages) == 0 {
		c.send(protocol.Message{Type: protocol.TypeSystem, Room: c.room, Content: "no history for room " + c.room})
		return
	}
	lines := make([]string, 0, len(messages)+1)
	lines = append(lines, "history for room "+c.room+":")
	for _, message := range messages {
		lines = append(lines, fmt.Sprintf("[%s] %s: %s", message.Timestamp.Local().Format("2006-01-02 15:04:05"), message.Username, message.Content))
	}
	c.send(protocol.Message{Type: protocol.TypeSystem, Room: c.room, Content: strings.Join(lines, "\n")})
}

func validateRoomName(name string) error {
	if name == "" {
		return fmt.Errorf("room name cannot be empty")
	}
	if strings.ContainsAny(name, " \t\r\n") {
		return fmt.Errorf("room name cannot contain whitespace")
	}
	return nil
}

func (c *client) send(message protocol.Message) {
	select {
	case c.out <- message:
	default:
		log.Printf("dropping message for %s: output buffer full", c.username)
	}
}
