#!/bin/bash

# Test script for Auth Service endpoints
# Make sure the service is running: make run

BASE_URL="http://localhost:8080"

echo "=== Testing Auth Service ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Health Check
echo -e "${YELLOW}1. Testing Health Check...${NC}"
HEALTH=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/health")
HTTP_STATUS=$(echo "$HEALTH" | grep "HTTP_STATUS" | cut -d: -f2)
if [ "$HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
    echo "$HEALTH" | grep -v "HTTP_STATUS"
else
    echo -e "${RED}✗ Health check failed (HTTP $HTTP_STATUS)${NC}"
fi
echo ""

# Test 2: Register User
echo -e "${YELLOW}2. Testing User Registration...${NC}"
REGISTER_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "SecurePass123!",
    "userType": "user"
  }')

HTTP_STATUS=$(echo "$REGISTER_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
if [ "$HTTP_STATUS" = "201" ]; then
    echo -e "${GREEN}✓ Registration successful${NC}"
    echo "$REGISTER_RESPONSE" | grep -v "HTTP_STATUS" | jq '.' 2>/dev/null || echo "$REGISTER_RESPONSE" | grep -v "HTTP_STATUS"
    
    # Extract access token if jq is available
    if command -v jq &> /dev/null; then
        ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | grep -v "HTTP_STATUS" | jq -r '.session.accessToken // empty')
        if [ -n "$ACCESS_TOKEN" ] && [ "$ACCESS_TOKEN" != "null" ]; then
            export ACCESS_TOKEN
            echo -e "${GREEN}  Access token extracted${NC}"
        fi
    fi
else
    echo -e "${RED}✗ Registration failed (HTTP $HTTP_STATUS)${NC}"
    echo "$REGISTER_RESPONSE" | grep -v "HTTP_STATUS"
fi
echo ""

# Test 3: Login (will fail if email not verified)
echo -e "${YELLOW}3. Testing User Login...${NC}"
LOGIN_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "SecurePass123!"
  }')

HTTP_STATUS=$(echo "$LOGIN_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
if [ "$HTTP_STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Login successful${NC}"
    echo "$LOGIN_RESPONSE" | grep -v "HTTP_STATUS" | jq '.' 2>/dev/null || echo "$LOGIN_RESPONSE" | grep -v "HTTP_STATUS"
    
    # Extract tokens if jq is available
    if command -v jq &> /dev/null; then
        ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -v "HTTP_STATUS" | jq -r '.session.accessToken // empty')
        REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -v "HTTP_STATUS" | jq -r '.session.refreshToken // empty')
        if [ -n "$ACCESS_TOKEN" ] && [ "$ACCESS_TOKEN" != "null" ]; then
            export ACCESS_TOKEN
            export REFRESH_TOKEN
            echo -e "${GREEN}  Tokens extracted${NC}"
        fi
    fi
elif [ "$HTTP_STATUS" = "403" ]; then
    echo -e "${YELLOW}⚠ Login requires email verification (expected for new users)${NC}"
    echo "$LOGIN_RESPONSE" | grep -v "HTTP_STATUS"
else
    echo -e "${RED}✗ Login failed (HTTP $HTTP_STATUS)${NC}"
    echo "$LOGIN_RESPONSE" | grep -v "HTTP_STATUS"
fi
echo ""

# Test 4: Logout (if we have a token)
if [ -n "$ACCESS_TOKEN" ] && [ "$ACCESS_TOKEN" != "null" ]; then
    echo -e "${YELLOW}4. Testing Logout...${NC}"
    LOGOUT_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL/v1/auth/logout" \
      -H "Authorization: Bearer $ACCESS_TOKEN")
    
    HTTP_STATUS=$(echo "$LOGOUT_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
    if [ "$HTTP_STATUS" = "204" ]; then
        echo -e "${GREEN}✓ Logout successful${NC}"
    else
        echo -e "${RED}✗ Logout failed (HTTP $HTTP_STATUS)${NC}"
        echo "$LOGOUT_RESPONSE" | grep -v "HTTP_STATUS"
    fi
    echo ""
else
    echo -e "${YELLOW}4. Skipping Logout (no access token available)${NC}"
    echo ""
fi

# Test 5: Password Reset Request
echo -e "${YELLOW}5. Testing Password Reset Request...${NC}"
RESET_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL/v1/auth/password/reset" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com"
  }')

HTTP_STATUS=$(echo "$RESET_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
if [ "$HTTP_STATUS" = "204" ]; then
    echo -e "${GREEN}✓ Password reset request successful${NC}"
else
    echo -e "${RED}✗ Password reset request failed (HTTP $HTTP_STATUS)${NC}"
    echo "$RESET_RESPONSE" | grep -v "HTTP_STATUS"
fi
echo ""

# Test 6: Invalid Login (wrong password)
echo -e "${YELLOW}6. Testing Invalid Login (wrong password)...${NC}"
INVALID_LOGIN=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "WrongPassword123!"
  }')

HTTP_STATUS=$(echo "$INVALID_LOGIN" | grep "HTTP_STATUS" | cut -d: -f2)
if [ "$HTTP_STATUS" = "401" ]; then
    echo -e "${GREEN}✓ Invalid login correctly rejected (HTTP 401)${NC}"
else
    echo -e "${RED}✗ Expected 401, got HTTP $HTTP_STATUS${NC}"
    echo "$INVALID_LOGIN" | grep -v "HTTP_STATUS"
fi
echo ""

echo "=== Test Summary ==="
echo "Check the results above for each test."
echo ""
echo "Note: Some tests may fail if:"
echo "  - Supabase is not running"
echo "  - Environment variables are not set"
echo "  - Email verification is required (Supabase default)"
echo ""
echo "To get Supabase credentials:"
echo "  1. Start Supabase: cd supabase && docker-compose up -d"
echo "  2. Open Studio: http://localhost:8000"
echo "  3. Go to Settings → API"
echo "  4. Copy credentials to .env file"

