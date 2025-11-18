# Testing Guide: Auth Service

This guide explains how to test the authentication service endpoints.

## Prerequisites

1. **Supabase running locally**:
   ```bash
   cd supabase
   docker-compose up -d
   ```

2. **Environment variables configured**:
   Create `.env` file in `services/auth/`:
   ```bash
   SUPABASE_URL=http://localhost:8000
   SUPABASE_ANON_KEY=<your-anon-key>
   SUPABASE_SERVICE_ROLE_KEY=<your-service-role-key>
   SUPABASE_JWT_SECRET=<your-jwt-secret>
   DATABASE_URL=postgresql://postgres:postgres@localhost:5432/postgres
   PORT=8080
   ENV=development
   ```

3. **Install dependencies**:
   ```bash
   cd services/auth
   go mod tidy
   ```

4. **Start the service**:
   ```bash
   make run
   # or
   go run ./cmd/auth
   ```

   Service will start on `http://localhost:8080`

## Testing Endpoints

### 1. Health Check

```bash
curl -X GET http://localhost:8080/health
```

**Expected Response**:
```json
{
  "status": "healthy",
  "timestamp": "",
  "version": "1.0.0"
}
```

---

### 2. User Registration

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "userType": "user"
  }'
```

**Expected Response** (201 Created):
```json
{
  "user": {
    "id": "uuid-here",
    "email": "test@example.com",
    "userType": "user"
  },
  "session": {
    "accessToken": "jwt-token-here",
    "refreshToken": "refresh-token-here",
    "expiresIn": 3600
  },
  "requiresVerification": true
}
```

**Test Streamer Registration**:
```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "streamer@example.com",
    "password": "SecurePass123!",
    "userType": "streamer"
  }'
```

**Test Validation Errors**:
```bash
# Invalid email
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalid-email",
    "password": "SecurePass123!",
    "userType": "user"
  }'

# Weak password
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "weak",
    "userType": "user"
  }'
```

---

### 3. Email Verification

**Note**: Supabase handles email verification via email links. This endpoint is informational.

```bash
curl -X GET "http://localhost:8080/v1/auth/verify?token=verification-token"
```

---

### 4. User Login

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

**Expected Response** (200 OK):
```json
{
  "user": {
    "id": "uuid-here",
    "email": "test@example.com",
    "userType": "user",
    "role": "user"
  },
  "session": {
    "accessToken": "jwt-token-here",
    "refreshToken": "refresh-token-here",
    "expiresIn": 3600
  }
}
```

**Test Invalid Credentials**:
```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "WrongPassword123!"
  }'
```

**Expected Response** (401 Unauthorized):
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid email or password"
  }
}
```

**Test Unverified Email**:
```bash
# Try to login before verifying email
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "unverified@example.com",
    "password": "SecurePass123!"
  }'
```

**Expected Response** (403 Forbidden):
```json
{
  "error": {
    "code": "EMAIL_NOT_VERIFIED",
    "message": "Email address not verified. Please check your email and verify your account."
  }
}
```

---

### 5. Refresh Token

```bash
# Save refresh token from login response
REFRESH_TOKEN="your-refresh-token-here"

curl -X POST http://localhost:8080/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{
    \"refreshToken\": \"$REFRESH_TOKEN\"
  }"
```

**Expected Response** (200 OK):
```json
{
  "session": {
    "accessToken": "new-jwt-token-here",
    "refreshToken": "new-refresh-token-here",
    "expiresIn": 3600
  }
}
```

---

### 6. Logout

```bash
# Save access token from login response
ACCESS_TOKEN="your-access-token-here"

curl -X POST http://localhost:8080/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

**Expected Response** (204 No Content)

**Test Without Token**:
```bash
curl -X POST http://localhost:8080/v1/auth/logout
```

**Expected Response** (401 Unauthorized):
```json
{
  "error": "missing authorization header"
}
```

---

### 7. Password Reset Request

```bash
curl -X POST http://localhost:8080/v1/auth/password/reset \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'
```

**Expected Response** (204 No Content)

**Note**: For security, this always returns success even if email doesn't exist.

---

### 8. Password Reset Confirmation

```bash
# Get reset token from email (Supabase sends it)
RESET_TOKEN="reset-token-from-email"

curl -X POST http://localhost:8080/v1/auth/password/reset/confirm \
  -H "Content-Type: application/json" \
  -d "{
    \"token\": \"$RESET_TOKEN\",
    \"password\": \"NewSecurePass123!\"
  }"
```

**Expected Response** (204 No Content)

**Test Invalid Token**:
```bash
curl -X POST http://localhost:8080/v1/auth/password/reset/confirm \
  -H "Content-Type: application/json" \
  -d '{
    "token": "invalid-token",
    "password": "NewSecurePass123!"
  }'
```

**Expected Response** (400 Bad Request):
```json
{
  "error": {
    "code": "BAD_REQUEST",
    "message": "Invalid or expired reset token. Please request a new password reset."
  }
}
```

---

### 9. OAuth Initiation

**For Streamers (Discord/Twitch)**:
```bash
curl -X GET "http://localhost:8080/v1/auth/oauth/discord?userType=streamer&redirect_uri=http://localhost:3000/callback"
```

**For Users (Google/Facebook)**:
```bash
curl -X GET "http://localhost:8080/v1/auth/oauth/google?userType=user&redirect_uri=http://localhost:3000/callback"
```

**Expected Response** (302 Redirect to OAuth provider)

**Test Invalid Provider for User Type**:
```bash
# Streamer trying to use Google (not allowed)
curl -X GET "http://localhost:8080/v1/auth/oauth/google?userType=streamer&redirect_uri=http://localhost:3000/callback"
```

**Expected Response** (400 Bad Request):
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Provider 'google' is not allowed for streamers. Use 'discord' or 'twitch'."
  }
}
```

---

### 10. OAuth Callback

**Note**: This is typically called by OAuth provider after user authorization.

```bash
curl -X GET "http://localhost:8080/v1/auth/oauth/discord/callback?code=oauth-code&state=state-value"
```

**Expected Response** (200 OK):
```json
{
  "user": {
    "id": "uuid-here",
    "email": "user@example.com",
    "userType": "streamer",
    "role": "streamer"
  },
  "session": {
    "accessToken": "jwt-token-here",
    "refreshToken": "refresh-token-here",
    "expiresIn": 3600
  }
}
```

---

## Complete Test Flow

### Full Registration → Login → Logout Flow

```bash
# 1. Register
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "SecurePass123!",
    "userType": "user"
  }')

echo "Registration: $REGISTER_RESPONSE"

# 2. Extract access token (requires jq or manual parsing)
# ACCESS_TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.session.accessToken')

# 3. Login (after email verification)
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "newuser@example.com",
    "password": "SecurePass123!"
  }')

echo "Login: $LOGIN_RESPONSE"

# 4. Extract tokens
ACCESS_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.session.accessToken')
REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.session.refreshToken')

# 5. Refresh token
REFRESH_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{
    \"refreshToken\": \"$REFRESH_TOKEN\"
  }")

echo "Refresh: $REFRESH_RESPONSE"

# 6. Logout
LOGOUT_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -w "\nHTTP Status: %{http_code}\n")

echo "Logout: $LOGOUT_RESPONSE"
```

---

## Testing with HTTPie

If you prefer HTTPie over curl:

```bash
# Install HTTPie
brew install httpie  # macOS
# or
pip install httpie

# Examples
http POST localhost:8080/v1/auth/register \
  email=test@example.com \
  password=SecurePass123! \
  userType=user

http POST localhost:8080/v1/auth/login \
  email=test@example.com \
  password=SecurePass123!

http POST localhost:8080/v1/auth/logout \
  Authorization:"Bearer $ACCESS_TOKEN"
```

---

## Testing with Postman

1. **Import Collection**: Create a new collection with all endpoints
2. **Set Environment Variables**:
   - `base_url`: `http://localhost:8080`
   - `access_token`: (set after login)
   - `refresh_token`: (set after login)
3. **Test Flow**:
   - Register → Login → Use token in protected endpoints → Logout

---

## Common Issues

### 1. Connection Refused
- **Issue**: `curl: (7) Failed to connect to localhost port 8080`
- **Solution**: Make sure the service is running (`make run`)

### 2. Supabase Connection Error
- **Issue**: `Failed to initialize Supabase client`
- **Solution**: 
  - Check Supabase is running: `docker-compose ps` in `supabase/`
  - Verify `SUPABASE_URL` and `SUPABASE_ANON_KEY` in `.env`

### 3. Database Connection Error
- **Issue**: `Failed to connect to database`
- **Solution**: 
  - Check PostgreSQL is running: `docker-compose ps` in `supabase/`
  - Verify `DATABASE_URL` in `.env`

### 4. Invalid JWT Secret
- **Issue**: `Invalid token` errors
- **Solution**: Verify `SUPABASE_JWT_SECRET` matches Supabase configuration

### 5. Email Not Verified
- **Issue**: Login returns 403 with "Email not verified"
- **Solution**: 
  - Check Supabase email settings (may need SMTP configuration)
  - Or manually verify email in Supabase Studio

---

## Next Steps

1. **Integration Tests**: Create automated tests using Go's testing package
2. **Load Testing**: Use tools like `k6` or `wrk` for performance testing
3. **Security Testing**: Test for common vulnerabilities (OWASP Top 10)
4. **End-to-End Tests**: Test complete user flows with frontend integration

