# GuestList - API Server

GuestList is a backend API built in Go for managing event ticketing, guest registration, and secure entry validation.

## Features

- **Event Management**: Create, list, edit, and delete events, as well as fetch detailed real-time hosting statistics.
- **Guest Ticketing**: Register guests, automatically generate unique secure ticket codes, and handle RSVP statuses.
- **Secure QR Check-in**: Authenticated endpoints to scan tickets and check in guests using custom cryptographic signature parameters.

---

## Technology Stack

- **Language**: Go 1.25+
- **Router**: Go Chi (v5)
- **Database Access**: PostgreSQL (via `pgx/v5` and `sqlc`)
- **Database Migrations**: Goose
- **Security**: JWT-based Authentication & Google OAuth2 integration

---

---

## Local Setup & Development

### 1. Prerequisites
- **Go**: Version `1.25.3` or higher.
- **PostgreSQL**: Local running database instance.
- **Goose**: Migration tool (optional, for manual migration commands).

### 2. Configuration
Copy a `.env.example` or create a `.env` file in the root directory. 

### 3. Database Migrations
Run database schema migrations using Goose:
```bash
goose up
```
*(Make sure the PostgreSQL database specified in `GOOSE_DBSTRING` has been created.)*

### 4. Running the Server
Start the HTTP API server:
```bash
go run ./cmd/...
```
The server will boot up by default on port `:8080` (or specified address in your configs).

---

