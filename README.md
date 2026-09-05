# Terminal Chat Application

A small multiplayer terminal chat application built with Go and raw TCP. Phase 2 adds multiple chat rooms and room-scoped broadcast messaging.

## Prerequisites

- Go 1.27 or newer

## Run

Start the server in one terminal:

```powershell
go run ./cmd/server -addr localhost:8080
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

Messages are newline-delimited JSON. A client registers with `{"type":"register","username":"alice"}`, joins with `{"type":"join_room","content":"lounge"}`, and sends chat with `{"type":"chat","content":"Hello"}`. The server broadcasts chat only to members of the sender's current room.

## Current scope

Phase 2 covers registration, room creation, joining, leaving, listing, and room-scoped messaging. Private messages, online-user listing, persistence, tests, Makefile targets, and graceful server shutdown will be added in later phases.
