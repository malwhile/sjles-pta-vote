# Authentication Setup Guide

## Environment Variables

The application now uses two simple environment variables for admin authentication:

### Required Variables

```bash
ADMIN_USER=<username>    # The admin username (plaintext)
ADMIN_PASS=<password>    # The admin password (plaintext at setup)
JWT_SECRET=<secret>      # Secret key for JWT signing (32+ characters)
```

## Setup Instructions

### 1. Generate a Strong JWT_SECRET

```bash
# Using OpenSSL (recommended)
export JWT_SECRET="$(openssl rand -base64 32)"

# Or use a strong random string (32+ characters)
export JWT_SECRET="abcdefghijklmnopqrstuvwxyz123456!@#$%^&"
```

### 2. Set Admin Credentials

```bash
# Choose a username and password
export ADMIN_USER="admin"
export ADMIN_PASS="your-secure-password-here"
```

### 3. Start the Server

```bash
go run ./server/main.go -setupdb
```

The system will:
1. Read `ADMIN_USER` and `ADMIN_PASS` from environment
2. Encrypt the password using `JWT_SECRET` as the encryption key
3. Store the encrypted password in memory
4. Use the encrypted version for all authentication checks

## How It Works

### Password Encryption

The admin password is encrypted using **AES-256-GCM** with the `JWT_SECRET`:

1. **Key Derivation**: JWT_SECRET → SHA256 → 32-byte encryption key
2. **Encryption**: Password is encrypted using AES-256-GCM with random nonce
3. **Storage**: Encrypted password (nonce + ciphertext) is stored in hex format

```go
// Internally, the flow is:
plainPassword := os.Getenv("ADMIN_PASS")
encryptedPassword := encryptPassword(plainPassword, jwtSecret)
// Later, during login:
decrypted := decryptPassword(encryptedPassword, jwtSecret)
if decrypted == providedPassword { ... }
```

### JWT Tokens

Once authenticated, admin users receive a JWT token:

```json
{
  "username": "admin",
  "exp": 1234567890,  // 24 hours from now
  "iat": 1234567000
}
```

Token is signed with `JWT_SECRET` and must be included in API requests:

```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/admin/new-poll
```

## Example Usage

### Development Setup

```bash
#!/bin/bash

# Generate strong secrets
export JWT_SECRET="$(openssl rand -base64 32)"
export ADMIN_USER="admin"
export ADMIN_PASS="dev-password-123"

# Start server
go run ./server/main.go -setupdb
```

### Production Setup

```bash
#!/bin/bash

# Use strong random secrets
export JWT_SECRET="$(openssl rand -base64 32)"
export ADMIN_USER="your_admin_username"
export ADMIN_PASS="very_strong_password_here"

# Build and run
go build -o pta-vote ./server
./pta-vote
```

### Login via API

```bash
# Request login
curl -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"dev-password-123"}'

# Response
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}

# Use token for admin endpoints
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  http://localhost:8080/api/admin/view-polls
```

## Security Considerations

### Current Approach
- ✅ Password encrypted at rest using AES-256-GCM
- ✅ Password only decrypted in memory for comparison
- ✅ JWT tokens signed with strong secret
- ✅ Tokens expire after 24 hours
- ✅ Failed logins are logged

### Limitations
- ⚠️ Password is plaintext in environment variables
- ⚠️ Single admin user (multi-user support would require database)
- ⚠️ No password rotation mechanism
- ⚠️ No rate limiting on login attempts

### Recommendations for Production
1. Use a secrets manager (AWS Secrets Manager, HashiCorp Vault)
2. Rotate secrets regularly
3. Monitor authentication logs
4. Use HTTPS only
5. Consider multi-factor authentication for future releases

## Troubleshooting

### "FATAL: ADMIN_USER and ADMIN_PASS environment variables not set"
**Solution**: Ensure both variables are exported:
```bash
export ADMIN_USER="admin"
export ADMIN_PASS="password"
go run ./server/main.go
```

### "Invalid username or password" login error
**Possible causes**:
- Typo in username or password
- Password contains special characters (try single quotes: `export ADMIN_PASS='pass@word'`)
- Environment variables not properly exported

**Debug**: Check that environment variables are set:
```bash
echo $ADMIN_USER
echo $ADMIN_PASS
echo $JWT_SECRET
```

### Token expiration / 401 Unauthorized
**Solution**: Login again to get a new token
```bash
curl -X POST http://localhost:8080/api/admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"..."}'
```

## Future Improvements

1. **Database-backed credentials**
   - Multiple admin users
   - Password history
   - Account lockout after failed attempts

2. **Better encryption**
   - Password pepper/salt
   - Bcrypt hashing instead of AES

3. **Enhanced security**
   - Multi-factor authentication
   - OAuth/OpenID Connect integration
   - Rate limiting on login attempts

4. **Audit logging**
   - Login attempts (success/failure)
   - Admin actions with timestamps
   - IP address logging
