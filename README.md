# Terminal Chat Application

A small multiplayer terminal chat application built with Go and raw TCP. Phase 1 provides the network foundation: newline-delimited JSON messages, concurrent TCP connections, username registration, and room-wide broadcast messaging in the default `general` room.

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

Each client asks for a unique username. Type normal text and press Enter to broadcast it to all connected clients. Enter `/quit` to disconnect.

Example:

```text
[alice@general] Hello Bob!
[bob@general] Hi Alice!
```

## Architecture

- `internal/protocol` contains the shared JSON message types and constants.
- `internal/server` accepts TCP connections, registers unique usernames, and uses one writer goroutine and buffered outgoing channel per client.
- `internal/client` reads terminal input while a separate goroutine receives server messages.
- `cmd/server` and `cmd/client` provide the command-line entry points.

Messages are newline-delimited JSON. A client registers with `{"type":"register","username":"alice"}` and sends chat with `{"type":"chat","content":"Hello"}`. The server broadcasts chat messages with the sender and room populated.

## Current scope

Phase 1 intentionally keeps the protocol and server small. Room management commands, private messages, online-user listing, persistence, tests, Makefile targets, and graceful server shutdown will be added in later phases.
