# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an online voting system for the SJLES PTA. It consists of a Go backend REST API with a React frontend.

## High-Level Architecture

### Backend (Go)
- **Layers**: HTTP handlers (main.go) → services/ → db/ + models/
- **Framework**: gorilla/mux for routing
- **Database**: SQLite (glebarez/go-sqlite driver)
- **Authentication**: JWT tokens (golang-jwt/jwt)
- **Server Port**: 8080

**Key packages**:
- `server/main.go`: HTTP handlers and route setup
- `server/services/`: Business logic (poll management, auth, members)
- `server/db/`: Database connection and schema
- `server/models/`: Data structures (Poll, Vote, Voter, Member)
- `server/config/`: Configuration from .env file
- `server/common/`: Shared utilities (error responses)

**Database schema**:
- `polls` table: poll questions, member/non-member vote counts, timestamps
- `voters` table: tracks who voted on each poll (composite key)
- `members` table: PTA members with emails and school year

### Frontend (React)
- **Build tool**: Create React App (react-scripts)
- **UI Library**: Material-UI (@mui/material)
- **Charts**: Recharts
- **Routing**: react-router
- **HTTP**: Axios

**Key pages**:
- `PollList.js`: Shows all available polls
- `PollDetails.js`: Shows detailed poll results
- `AdminLogin.js`: Admin authentication
- `AdminMembers.js`: Member management
- `AdminCreateVote.js`: Create new polls

The built React app is served from the Go server at the root path, with API requests to `/api/*` endpoints.

## Development Commands

### Backend (Go)

**Run server** (from repo root):
```bash
go run ./server/main.go
```

**Run server with sample data** (populates 3 test polls):
```bash
go run ./server/main.go -setupdb
```

**Build binary**:
```bash
go build -o pta-vote ./server
```

**Run tests**:
```bash
go test ./...
```

**Run single test** (e.g., db tests):
```bash
go test ./server/db -v
```

**Lint** (using go vet):
```bash
go vet ./...
```

### Frontend (React)

All npm commands must be run from the `client/` directory:

**Start dev server** (runs on port 3000, proxies API to port 8080):
```bash
cd client && npm start
```

**Build for production**:
```bash
cd client && npm build
```

**Run React tests**:
```bash
cd client && npm test
```

### Full Stack Development

**Setup and run both (in separate terminals)**:
1. Terminal 1: `go run ./server/main.go -setupdb`
2. Terminal 2: `cd client && npm start`

The React dev server will proxy `/api` requests to the Go server.

## Key Architectural Details

### API Endpoints
- `POST /api/admin/login`: Admin authentication, returns JWT token
- `POST /api/admin/new-poll`: Create new poll (requires auth)
- `GET /api/admin/view-polls`: List all polls (requires auth)
- `POST /api/vote`: Submit a vote (body: {pollId, email, vote})
- `GET /api/polls/{id}`: Get poll details with vote counts
- `POST /api/stats`: Get poll statistics
- `POST /api/admin/members`: Manage PTA members (requires auth)
- `GET /api/admin/members/view`: View members (requires auth)

### Configuration
Configuration is read from `.env` file (auto-created on first run):
```
db_path="pta_vote.db"
```

The config package maintains a singleton `Config` struct with database path and Redis settings (Redis fields exist but not yet used).

### Testing Approach
- Go tests use testify assertions and temporary SQLite databases
- Tests create their own temp .db files and set config before running
- db_test.go shows the pattern: create temp db, set config, run test, cleanup

### Git Branches
Currently on branch `6-poll-detail` which adds poll detail pages and pass/fail columns. Main branch is `main`.
