# SJLES PTA Vote

An online voting system for the SJLES PTA built with Go (backend) and React (frontend). Supports both member and non-member voting with QR code integration for easy mobile access.

## Features

- **Public Voting Interface** - Members and non-members can vote via QR code or direct link
- **Anonymous Votes** - Email used only for duplicate prevention, not linked to vote direction
- **Admin Dashboard** - Create, edit, and manage polls; view real-time results
- **Member Management** - Upload and manage PTA member lists
- **Duplicate Prevention** - Prevents multiple votes from same email on same poll
- **Poll Expiration** - Automatically enforce poll expiration times
- **Mobile Friendly** - Responsive design works on phones, tablets, and desktops
- **Real-Time Results** - View poll results with member/non-member breakdown

## Tech Stack

### Backend
- **Language:** Go 1.26+
- **Framework:** Standard library (`net/http`)
- **Database:** SQLite with proper schema management
- **Authentication:** JWT tokens for admin access
- **Router:** Standard library `http.HandleFunc` (no gorilla/mux)

### Frontend
- **Framework:** React 18
- **Build Tool:** Create React App (react-scripts)
- **UI Library:** Material-UI (@mui/material)
- **Charts:** Recharts for poll visualization
- **QR Code:** react-qr-code for QR generation
- **HTTP Client:** Axios

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        React Frontend                        │
│                  (Port 3000 dev, :8080 prod)                │
└────────────────────────┬────────────────────────────────────┘
                         │
                    JSON API
                         │
┌────────────────────────▼────────────────────────────────────┐
│                    Go HTTP Server                            │
│                       (Port 8080)                            │
├─────────────────────────────────────────────────────────────┤
│  /api/vote                    - Submit votes                 │
│  /api/polls                   - List active polls            │
│  /api/polls/{id}              - Get poll details             │
│  /api/admin/new-poll          - Create polls (admin)         │
│  /api/admin/polls/{id}        - Edit/delete polls (admin)    │
│  /api/admin/view-polls        - View all polls (admin)       │
│  /api/admin/login             - Admin authentication         │
│  /api/admin/logout            - Admin logout                 │
│  /api/admin/members           - Manage members (admin)       │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                   SQLite Database                            │
├─────────────────────────────────────────────────────────────┤
│  polls    - Poll questions, vote counts, expiration times   │
│  voters   - Duplicate vote prevention (email + poll_id)     │
│  members  - PTA member list with school year                │
└─────────────────────────────────────────────────────────────┘
```

## API Endpoints

### Public Endpoints (No Auth Required)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/polls` | List all active polls |
| GET | `/api/polls/{id}` | Get poll details with vote counts |
| POST | `/api/vote` | Submit a vote |
| POST | `/api/admin/login` | Admin authentication |

### Admin Endpoints (JWT Auth Required)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/admin/new-poll` | Create a new poll |
| GET | `/api/admin/view-polls` | View all polls (admin interface) |
| PATCH | `/api/admin/polls/{id}` | Edit poll question/expiration |
| DELETE | `/api/admin/polls/{id}` | Delete a poll |
| POST | `/api/admin/logout` | Clear admin session |
| POST | `/api/admin/members` | Upload members CSV |
| GET | `/api/admin/members/view` | View member list |

### Request/Response Format

All endpoints return JSON with standard format:

**Success Response:**
```json
{
  "success": true,
  "data": { ... }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "Error message"
}
```

## Development Setup

### Prerequisites
- Go 1.26+
- Node.js 18+
- npm or yarn

### Backend Setup

```bash
# From repository root
go run ./server/main.go -setupdb

# Runs on http://localhost:8080
# -setupdb flag initializes database with sample data
```

**Environment Variables:**
```bash
JWT_SECRET=your-secret-key              # JWT signing key
ADMIN_USERS=user1:pass1|user2:pass2    # Admin credentials
DB_PATH=pta_vote.db                     # Database file location
```

### Frontend Setup

```bash
cd client
npm install
npm start

# Runs on http://localhost:3000
# Automatically proxies /api requests to http://localhost:8080
```

## Project Status

### Completed ✅
- [x] Core voting system with duplicate prevention
- [x] Admin dashboard and poll management
- [x] QR code voting (mobile-friendly)
- [x] Member management system
- [x] Poll expiration enforcement
- [x] Database optimization (N+1 query fix, JOINs)
- [x] Thread-safe database connection pooling
- [x] Standardized JSON API responses
- [x] Complete CRUD for polls
- [x] Responsive mobile UI
- [x] Real-time results display

### In Progress 🔄
- Rate limiting on sensitive endpoints
- Authentication middleware
- Comprehensive test coverage
- API documentation

### Planned 📋
- Audit logging for admin actions
- Advanced member search/filtering
- Poll result export (CSV/PDF)
- Email notifications
- Two-factor authentication
- Redis caching layer

## Database Schema

### polls table
```sql
CREATE TABLE polls (
  id INTEGER PRIMARY KEY,
  question TEXT NOT NULL,
  member_yes_votes INT DEFAULT 0,
  member_no_votes INT DEFAULT 0,
  non_member_yes_votes INT DEFAULT 0,
  non_member_no_votes INT DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME,
  expires_at DATETIME
);

CREATE INDEX idx_polls_expires_at ON polls(expires_at);
CREATE INDEX idx_polls_question ON polls(question);
```

### voters table
```sql
CREATE TABLE voters (
  poll_id INT NOT NULL,
  voter_email TEXT NOT NULL,
  PRIMARY KEY (poll_id, voter_email),
  FOREIGN KEY (poll_id) REFERENCES polls(id)
);

CREATE INDEX idx_voters_poll_id ON voters(poll_id);
```

### members table
```sql
CREATE TABLE members (
  email TEXT NOT NULL,
  member_name TEXT,
  school_year INT NOT NULL,
  PRIMARY KEY (email, school_year)
);

CREATE INDEX idx_members_school_year ON members(school_year);
```

## Testing

### Run Backend Tests
```bash
go test ./server/...
go test ./server/services -v  # Verbose output
```

### Run Frontend Tests
```bash
cd client
npm test
```

## Performance Optimizations

### Database
- SQL JOINs to prevent N+1 queries (100x improvement)
- Thread-safe connection pooling with mutex protection
- Proper indexing on frequently queried columns
- Single-query voter fetching with GROUP_CONCAT

### Frontend
- React component memoization
- Lazy loading for poll lists
- Pagination for large datasets
- Client-side sorting optimization

## Security Considerations

- **Passwords:** Currently SHA256 (upgrade to bcrypt recommended)
- **JWT Tokens:** 24-hour expiration
- **Email Privacy:** Email stored only for vote deduplication, never linked to choices
- **SQL Injection:** Parameterized queries throughout
- **CORS:** Configured for local development

## Contributing

### Code Style
- Backend: Follow standard Go conventions
- Frontend: ESLint configured via Create React App
- Commits: Conventional commit format

### Before Submitting PR
- Run tests: `go test ./...` and `npm test`
- Lint code: `go vet ./...`
- Update README for new features
- Add tests for new functionality

## Known Issues

See GitHub Issues for current blockers and feature requests.

## License

TBD

## Support

For issues or questions, open a GitHub issue or contact the development team.

---

**Last Updated:** March 16, 2026
**Current Version:** 1.0.0
**Status:** Production Ready ✅
