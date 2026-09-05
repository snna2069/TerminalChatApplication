# Terminal Chat Application

A small multiplayer terminal chat application built with Go and raw TCP. Phase 4 adds JSON Lines message persistence and room history.

## Prerequisites

- Go 1.27 or newer

## Run

Start the server in one terminal:

```powershell
go run ./cmd/server -addr localhost:8080
# Optional custom history file:
go run ./cmd/server -addr localhost:8080 -data data/messages.jsonl
```

Start two or more clients in separate terminals:

```powershell
go run ./cmd/client -addr localhost:8080
```

Each client asks for a unique username and starts in `general`. Type normal text and press Enter to broadcast it to members of the current room. Enter `/quit` to disconnect.

Room commands:

```text
/rooms          List available rooms
/create lounge  Create and join a room
/join lounge    Join an existing room
/leave          Return to general
/users          List online users and their rooms
/msg bob Hi!    Send a private message
/history        Show recent history for the current room
/history 50     Show up to 50 recent messages
/help           Show command usage
/quit           Disconnect
```

Example:

```text
[alice@general] Hello Bob!
[bob@general] Hi Alice!
```

## Architecture

- `internal/protocol` contains the shared JSON message types and constants.
- `internal/server` accepts TCP connections, tracks unique users and rooms, and uses one writer goroutine and buffered outgoing channel per client.
- `internal/client` reads terminal input while a separate goroutine receives server messages.
- `cmd/server` and `cmd/client` provide the command-line entry points.

Messages are newline-delimited JSON. A client registers with `{"type":"register","username":"alice"}`, joins with `{"type":"join_room","content":"lounge"}`, sends chat with `{"type":"chat","content":"Hello"}`, requests history with `{"type":"history","limit":20}`, and sends private messages with `{"type":"private_message","target":"bob","content":"Hi"}`. The server broadcasts chat only to members of the sender's current room and sends private messages only to the sender and recipient.

## Message persistence

The server stores room chat messages as one JSON object per line in `data/messages.jsonl` by default. Each record contains a UTC timestamp, username, room, and message content. The storage interface is isolated in `internal/store`, so the JSON Lines backend can be replaced later.

## Current scope

Phase 4 covers private messages, online-user listing, complete command parsing, JSON Lines message persistence, and room history. Makefile targets, integration tests, and graceful server shutdown will be added in later phases.
