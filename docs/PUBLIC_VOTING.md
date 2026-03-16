# Public Voting System

## Overview

The SJLES PTA voting system is **completely open to the public**. Any visitor can:
- ✅ View all active polls
- ✅ View poll details and results
- ✅ Vote on polls
- ✅ Share voting links (QR codes)

**No authentication is required** to participate in voting.

## Security Model

Since voting is public, security relies on:

### 1. Rate Limiting (Primary Control)
Prevents abuse by limiting requests per IP address:

| Endpoint | Limit | Window | Purpose |
|----------|-------|--------|---------|
| `POST /api/vote` | 100 votes | 1 hour | Prevent vote flooding |
| `GET /api/polls` | 300 views | 1 minute | Prevent poll list spam |
| `GET /api/polls/{id}` | 300 views | 1 minute | Prevent poll detail spam |

### 2. Duplicate Vote Prevention
- Email address used to track voters per poll
- Same email cannot vote twice on same poll
- Email is **not linked** to vote choice (anonymous)

### 3. Poll Expiration
- Admin sets expiration time when creating poll
- Voting disabled after expiration
- Backend enforces: returns `403 Forbidden` if expired

### 4. Database Indexes
- Indexed voter lookups (fast duplicate detection)
- Indexed expiration checks
- Prevents denial-of-service via slow queries

## API Endpoints

### Public Endpoints (No Authentication)

#### List All Polls
```bash
GET /api/polls
Response: Array of poll objects
Rate limit: 300 requests/minute per IP
```

#### Get Poll Details
```bash
GET /api/polls/{id}
Response: {
  "id": 1,
  "question": "Should we increase budget?",
  "member_yes": 10,
  "member_no": 5,
  "non_member_yes": 3,
  "non_member_no": 2,
  "expires_at": "2026-03-20T10:00:00Z"
}
Rate limit: 300 requests/minute per IP
```

#### Submit Vote
```bash
POST /api/vote
Body: {
  "poll_id": 1,
  "email": "voter@example.com",
  "vote": true  // true = yes, false = no
}

Response (200 OK): {"success": true}
Response (409 Conflict): {"success": false, "error": "Already voted"}
Response (403 Forbidden): {"success": false, "error": "Poll has expired"}
Response (429 Too Many Requests): {"success": false, "error": "Too many requests..."}

Rate limit: 100 votes/hour per IP
```

### Admin Endpoints (Authentication Required)

#### Admin Login
```bash
POST /api/admin/login
Body: {"username": "admin", "password": "..."}
Response: {"success": true, "token": "eyJhbGc..."}
```

#### Create Poll
```bash
POST /api/admin/new-poll
Headers: {"Authorization": "Bearer <token>"}
Body: {"question": "Question?", "duration_hours": 24}
```

#### View All Polls (Admin Dashboard)
```bash
POST /api/admin/view-polls
Headers: {"Authorization": "Bearer <token>"}
Body: {"poll_id": 1}  # Optional - get specific poll
```

#### Edit Poll
```bash
PATCH /api/admin/polls/{id}
Headers: {"Authorization": "Bearer <token>"}
Body: {"question": "New question?", "duration_hours": 48}
```

#### Delete Poll
```bash
DELETE /api/admin/polls/{id}
Headers: {"Authorization": "Bearer <token>"}
```

## Voting Flow

### 1. QR Code / Share Link
Admin generates QR code or share link:
```
https://pta-vote.com/vote/42
```

### 2. Visitor Opens Link
- Mobile-friendly voting page loads
- Poll question and description displayed
- Email input field shown

### 3. Voter Submits
- Email: `voter@example.com`
- Vote: Yes or No
- Frontend validates inputs
- Backend checks:
  - ✓ Poll exists
  - ✓ Poll not expired
  - ✓ Email hasn't voted on this poll
  - ✓ Not rate limited

### 4. Vote Recorded
- Vote stored anonymously
- Email tracked for duplicate prevention
- Poll results updated in real-time
- Voter sees confirmation

## Rate Limiting Details

### Algorithm: Token Bucket
- Per IP address tracking
- Automatic reset after time window
- Old entries cleaned up every 5-10 minutes

### Examples

**Normal voting:**
```bash
# Visitor votes on 3 polls = 3 requests
# Within rate limit (100/hour) ✓
```

**Potential attack:**
```bash
# Bot submits 101 votes in 1 hour from same IP
# Request 101 rejected: 429 Too Many Requests ✗
```

**Different IPs:**
```bash
# 100 different IPs, each vote once = 100 requests
# All allowed (per-IP limits) ✓
```

### Proxy Detection
Rate limiter checks headers for proxies:
1. `X-Forwarded-For` (nginx, load balancers)
2. `X-Real-IP` (nginx)
3. `RemoteAddr` (direct connection)

## Privacy & Security Guarantees

✅ **Vote Privacy**
- Email stored but not linked to vote choice
- Admins cannot see who voted for what
- Anonymous voting for all

✅ **Duplicate Prevention**
- Email-based tracking per poll
- Cannot vote twice on same poll
- Different polls can use same email

✅ **Data Integrity**
- Poll expiration enforced
- Database indexes prevent slow queries
- JWT authentication for admin actions

✅ **Availability**
- Rate limiting prevents DoS attacks
- Efficient queries for quick response
- Handles 1000+ concurrent voters

## Deployment Considerations

### For Small Events (50 voters)
- Default rate limits are sufficient
- No special scaling needed
- Single server/database OK

### For Medium Events (500+ voters)
- Monitor rate limit logs
- Consider increasing limits if legitimate traffic blocked
- Add reverse proxy (nginx) for IP tracking

### For Large Events (1000+ voters)
- Scale to multiple servers
- Use load balancer with sticky sessions
- Consider Redis caching for polls
- Increase rate limits as needed

## Configuration

### Change Rate Limits
Edit `middleware/rate_limit.go`:

```go
// Vote endpoint: 200 votes per hour per IP
VoteRateLimiter = NewRateLimiter(200, time.Hour, 10*time.Minute)

// Poll view endpoint: 600 views per minute per IP
PollViewRateLimiter = NewRateLimiter(600, time.Minute, 5*time.Minute)
```

### Monitor Rate Limiting
Check logs for rate limit events:
```bash
# See who hit rate limits
grep "rate limit exceeded" logs/server.log
```

## Future Enhancements

1. **CAPTCHA Integration**
   - Add CAPTCHA to voting form
   - Prevent automated attacks

2. **Email Verification**
   - Send confirmation link to email
   - Verify ownership before counting vote

3. **Member-Only Polls**
   - Create polls restricted to members
   - Upload member list for validation

4. **Admin Dashboard**
   - Real-time vote tracking
   - IP-based voting patterns
   - Fraud detection

5. **Advanced Rate Limiting**
   - Geographic blocking
   - Device fingerprinting
   - Behavior analysis
