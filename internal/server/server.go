package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/snna2069/TerminalChatApplication/internal/protocol"
)

type Server struct {
	addr    string
	mu      sync.RWMutex
	clients map[*client]struct{}
}

type client struct {
	conn     net.Conn
	username string
	out      chan protocol.Message
}

func New(addr string) *Server {
	return &Server{addr: addr, clients: make(map[*client]struct{})}
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
			c.send(protocol.Message{Type: protocol.TypeSystem, Room: protocol.DefaultRoom, Content: fmt.Sprintf("connected as %s; joined room %s", c.username, protocol.DefaultRoom)})
			s.broadcast(protocol.Message{Type: protocol.TypeSystem, Content: fmt.Sprintf("%s joined the chat", c.username)}, c)
			continue
		}

		if message.Type == protocol.TypeChat {
			content := strings.TrimSpace(message.Content)
			if content == "" {
				c.send(protocol.Message{Type: protocol.TypeError, Content: "message cannot be empty"})
				continue
			}
			s.broadcast(protocol.Message{Type: protocol.TypeChat, Username: c.username, Room: protocol.DefaultRoom, Content: content}, nil)
			continue
		}
		c.send(protocol.Message{Type: protocol.TypeError, Content: "unsupported message type"})
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
	s.clients[c] = struct{}{}
	return nil
}

func (s *Server) removeClient(c *client) {
	s.mu.Lock()
	delete(s.clients, c)
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

func (c *client) send(message protocol.Message) {
	select {
	case c.out <- message:
	default:
		log.Printf("dropping message for %s: output buffer full", c.username)
	}
}
