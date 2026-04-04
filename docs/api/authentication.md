# Authentication Guide

This document describes the authentication mechanisms used in the VN Stock Analysis System.

## Authentication Methods

The system supports two authentication methods:

1. **JWT Tokens** - For user authentication (end-users)
2. **API Keys** - For service-to-service communication

---

## JWT Token Authentication

### Overview

JWT (JSON Web Token) authentication is used for user-based access to the API. Tokens are signed using HMAC-SHA256 and include user information and permissions.

### Configuration

Add the following to your `.env` file:

```bash
# Enable authentication
ENABLE_AUTH=true

# JWT secret key (change in production!)
JWT_SECRET_KEY=your-secure-secret-key-min-32-chars

# Token expiration (hours)
JWT_EXPIRATION_HOURS=24
```

### Obtaining a Token

Tokens are obtained through your authentication system. For development, you can generate a test token:

```bash
# Example: Generate a test token (requires implementing a /auth/login endpoint)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "password": "password"}'
```

### Using the Token

Include the JWT token in the `Authorization` header:

```bash
curl -X GET http://localhost:8080/api/v1/technical/VNM \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Token Structure

```json
{
  "user_id": "user-123",
  "email": "user@example.com",
  "role": "analyst",
  "scopes": ["read:technical", "read:sentiment"],
  "service": "",
  "exp": 1706547200,
  "iat": 1706460800,
  "iss": "vnstock-api"
}
```

### Scopes

| Scope | Description |
|-------|-------------|
| `read:technical` | Read technical analysis data |
| `read:sentiment` | Read sentiment analysis data |
| `write:analysis` | Create new analysis |
| `admin:*` | Full admin access |

---

## API Key Authentication

### Overview

API keys are used for service-to-service communication, such as when the Go API Gateway calls the Python Sentiment Service.

### Configuration

Add the following to your `.env` file:

```bash
# API Key header name
API_KEY_HEADER=X-API-Key

# API keys (format: key:value,key:value)
API_KEYS=service-key-1:sentiment-service,service-key-2:external-api
```

### Using API Keys

Include the API key in the `X-API-Key` header:

```bash
curl -X POST http://localhost:8000/api/v1/analyze \
  -H "Content-Type: application/json" \
  -H "X-API-Key: service-key-1" \
  -d '{"text": "Cổ phiếu VNM tăng mạnh"}'
```

---

## Error Responses

### 401 Unauthorized

```json
{
  "error": "missing authorization header",
  "code": "UNAUTHORIZED"
}
```

```json
{
  "error": "invalid token",
  "code": "TOKEN_INVALID"
}
```

```json
{
  "error": "token has expired",
  "code": "TOKEN_INVALID"
}
```

```json
{
  "error": "missing API key",
  "code": "MISSING_API_KEY"
}
```

```json
{
  "error": "invalid API key",
  "code": "INVALID_API_KEY"
}
```

### 403 Forbidden

```json
{
  "error": "insufficient permissions",
  "code": "FORBIDDEN",
  "required_scope": "read:technical"
}
```

---

## Security Best Practices

1. **Use strong JWT secret keys** - At least 32 characters, randomly generated
2. **Rotate API keys regularly** - Change keys periodically
3. **Use HTTPS in production** - Never transmit tokens over unencrypted connections
4. **Set appropriate token expiration** - Shorter is more secure
5. **Implement IP allowlisting** - For API keys in production
6. **Log authentication failures** - Monitor for brute force attempts

---

## Development vs Production

### Development

In development, authentication can be disabled:

```bash
ENABLE_AUTH=false
```

This allows access to all endpoints without authentication.

### Production

In production, always enable authentication:

```bash
ENABLE_AUTH=true
JWT_SECRET_KEY=<32+ character random string>
API_KEYS=<your-service-keys>
```

---

## Example: Generating a Test JWT Token

You can generate a test JWT token using the middleware package:

```go
package main

import (
    "fmt"
    "log"

    "vnstock-hybrid/internal/middleware"
)

func main() {
    claims := &middleware.Claims{
        UserID: "test-user",
        Email:  "test@example.com",
        Role:   "analyst",
        Scopes: []string{"read:technical", "read:sentiment"},
    }

    token, err := middleware.GenerateToken(claims, "your-secret-key", 24)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Token:", token)
}
```
