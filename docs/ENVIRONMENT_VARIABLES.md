# Environment Variables

Central Logs supports comprehensive environment variable configuration to override values from `config.yaml`. This is especially useful for:

- 🐳 **Docker deployments** - Pass secrets without modifying config files
- ☁️ **Cloud platforms** - Use platform environment variables (Heroku, AWS, etc.)
- 🔒 **Security** - Keep sensitive data out of version control
- 🔄 **CI/CD** - Configure different environments easily

## Dual Format Support

All environment variables support **both formats**:

1. **Direct format** (recommended): `VARIABLE_NAME`
   ```bash
   export SERVER_PORT=8080
   export DATABASE_PATH=/var/lib/db.sqlite
   ```

2. **Prefixed format** (backward compatible): `CL_VARIABLE_NAME`
   ```bash
   export CL_SERVER_PORT=8080
   export CL_DATABASE_PATH=/var/lib/db.sqlite
   ```

**Priority**: If both are set, the direct format (without `CL_` prefix) takes precedence.

## Quick Reference

### Server Configuration

```bash
# Port number (default: 3000)
export SERVER_PORT=8080

# Environment: development, production (default: development)
export SERVER_ENV=production
```

### Database Configuration

```bash
# SQLite database file path (default: ./data/central-logs.db)
export DATABASE_PATH=/var/lib/central-logs/db.sqlite
```

### Redis Configuration

```bash
# Redis connection URL (default: redis://localhost:6379)
export REDIS_URL=redis://redis-server:6379
export REDIS_URL=redis://:password@redis-server:6379/0
```

### JWT Authentication

```bash
# JWT secret key (CHANGE IN PRODUCTION!)
export JWT_SECRET=your-super-secret-key-here

# JWT token expiry (default: 24h)
export JWT_EXPIRY=48h
```

### Web Push (VAPID)

```bash
# VAPID public key for web push notifications
export VAPID_PUBLIC_KEY="BCPvbPUl5BxOmMbJGV7Yorn_JB-EYfZM..."

# VAPID private key (KEEP SECRET!)
export VAPID_PRIVATE_KEY="4dERHXcVs9WYkT8PtwFMid7mpMf_z1..."

# VAPID subject (mailto or https URL)
export VAPID_SUBJECT=mailto:admin@example.com
```

### Admin User

```bash
# Initial admin username (default: admin)
export ADMIN_USERNAME=superadmin

# Initial admin password (CHANGE IN PRODUCTION!)
export ADMIN_PASSWORD=SecurePassword123!
```

### Rate Limiting

```bash
# API requests per minute (default: 1000)
export RATE_LIMIT_API_REQUESTS_PER_MINUTE=5000

# Telegram messages per minute (default: 20)
export RATE_LIMIT_TELEGRAM_MESSAGES_PER_MINUTE=30

# Discord messages per minute (default: 30)
export RATE_LIMIT_DISCORD_MESSAGES_PER_MINUTE=50

# Push notifications per minute (default: 60)
export RATE_LIMIT_PUSH_MESSAGES_PER_MINUTE=100
```

### WebSocket Configuration

```bash
# Enable/disable WebSocket (default: true)
export WEBSOCKET_ENABLED=true

# Ping interval (default: 30s)
export WEBSOCKET_PING_INTERVAL=60s

# Pong timeout (default: 10s)
export WEBSOCKET_PONG_TIMEOUT=15s

# Max message size in bytes (default: 512)
export WEBSOCKET_MAX_MESSAGE_SIZE=1024

# Read buffer size (default: 1024)
export WEBSOCKET_READ_BUFFER_SIZE=2048

# Write buffer size (default: 1024)
export WEBSOCKET_WRITE_BUFFER_SIZE=2048
```

### Retention Policy

```bash
# Enable retention cleanup (default: true)
export RETENTION_ENABLED=true

# Default retention max age (default: 30d)
export RETENTION_DEFAULT_MAX_AGE=90d

# Default retention max count (default: 100000)
export RETENTION_DEFAULT_MAX_COUNT=1000000

# Enable cleanup job (default: true)
export RETENTION_CLEANUP_ENABLED=true

# Cleanup schedule (cron format, default: 0 2 * * *)
export RETENTION_CLEANUP_SCHEDULE="0 3 * * *"

# Cleanup batch size (default: 1000)
export RETENTION_CLEANUP_BATCH_SIZE=5000
```

## Usage Examples

### Docker Compose

```yaml
version: '3.8'

services:
  central-logs:
    image: ghcr.io/pandeptwidyaop/central-logs:latest
    ports:
      - "8080:8080"
    environment:
      # Server
      - SERVER_PORT=8080
      - SERVER_ENV=production

      # Database
      - DATABASE_PATH=/data/central-logs.db

      # Redis
      - REDIS_URL=redis://redis:6379

      # JWT (use secrets in production!)
      - JWT_SECRET=${JWT_SECRET}

      # Admin
      - ADMIN_PASSWORD=${ADMIN_PASSWORD}
    volumes:
      - ./data:/data
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
```

### Kubernetes ConfigMap & Secret

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: central-logs-config
data:
  SERVER_PORT: "3000"
  SERVER_ENV: "production"
  DATABASE_PATH: "/data/central-logs.db"
  REDIS_URL: "redis://redis-service:6379"
---
apiVersion: v1
kind: Secret
metadata:
  name: central-logs-secrets
type: Opaque
stringData:
  JWT_SECRET: "your-secret-key"
  ADMIN_PASSWORD: "secure-password"
  VAPID_PRIVATE_KEY: "your-private-key"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: central-logs
spec:
  replicas: 1
  selector:
    matchLabels:
      app: central-logs
  template:
    metadata:
      labels:
        app: central-logs
    spec:
      containers:
      - name: central-logs
        image: ghcr.io/pandeptwidyaop/central-logs:latest
        envFrom:
        - configMapRef:
            name: central-logs-config
        - secretRef:
            name: central-logs-secrets
```

### Systemd Service

```ini
[Unit]
Description=Central Logs
After=network.target

[Service]
Type=simple
User=central-logs
WorkingDirectory=/opt/central-logs
ExecStart=/opt/central-logs/bin/server

# Environment variables
Environment="SERVER_PORT=3000"
Environment="SERVER_ENV=production"
Environment="DATABASE_PATH=/var/lib/central-logs/db.sqlite"
EnvironmentFile=/etc/central-logs/env

Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

File `/etc/central-logs/env`:
```bash
SERVER_PORT=3000
JWT_SECRET=your-secret-key
ADMIN_PASSWORD=secure-password
REDIS_URL=redis://localhost:6379
```

### .env File (for development)

Create a `.env` file in your project root:

```bash
# Server
SERVER_PORT=3000
SERVER_ENV=development

# Database
DATABASE_PATH=./data/dev.db

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=dev-secret-key
JWT_EXPIRY=24h

# Admin
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123

# VAPID (get from https://vapidkeys.com/)
VAPID_PUBLIC_KEY="your-public-key"
VAPID_PRIVATE_KEY="your-private-key"
VAPID_SUBJECT=mailto:dev@localhost
```

Then use [direnv](https://direnv.net/) or source it manually:
```bash
source .env
./bin/server
```

## Viewing All Supported Variables

Run the env help command to see all supported variables:

```bash
go run cmd/envhelp/main.go
```

Or after building:

```bash
./bin/envhelp
```

## Security Best Practices

### ⚠️ Sensitive Variables

Never commit these to version control:
- `JWT_SECRET`
- `ADMIN_PASSWORD`
- `VAPID_PRIVATE_KEY`
- `REDIS_URL` (if contains password)

### ✅ Recommended Setup

1. **Development**: Use `config.yaml` with non-sensitive defaults
2. **Production**: Override sensitive values via environment variables
3. **CI/CD**: Use platform secrets (GitHub Secrets, GitLab CI Variables, etc.)
4. **Docker**: Use Docker secrets or environment files
5. **Kubernetes**: Use Secrets for sensitive data, ConfigMaps for non-sensitive

### Example .gitignore

```gitignore
# Sensitive files
.env
.env.local
.env.production
config.production.yaml

# Allow example files
!.env.example
```

## Type Conversion

Environment variables are automatically converted to the correct type:

- **string**: Direct value (e.g., `SERVER_ENV=production`)
- **int**: Parsed as integer (e.g., `SERVER_PORT=8080`)
- **bool**: Parsed as boolean (e.g., `WEBSOCKET_ENABLED=true` or `false`)
- **duration**: Parsed as Go duration (e.g., `JWT_EXPIRY=24h`, `WEBSOCKET_PING_INTERVAL=30s`)

## Validation

The application will:
- ✅ Use default values if environment variable is not set
- ✅ Override config.yaml if environment variable is set
- ⚠️ Print warning if type conversion fails
- ⚠️ Use default value if conversion fails

## Troubleshooting

### Variable not working?

1. **Check the variable name**: Use `go run cmd/envhelp/main.go` to see exact names
2. **Check for typos**: Environment variable names are case-sensitive
3. **Check the value**: Some variables require specific formats (e.g., URLs, durations)
4. **Check if it's set**: Run `echo $VARIABLE_NAME` in your shell

### Example debugging

```bash
# Set variable
export SERVER_PORT=8080

# Verify it's set
echo $SERVER_PORT

# Run with debug
./bin/server
# Check startup logs for "Port: 8080"
```

### Priority doesn't work as expected?

Remember the priority order:
1. Direct format (e.g., `SERVER_PORT`)
2. Prefixed format (e.g., `CL_SERVER_PORT`)
3. `config.yaml`
4. Default values

If you set both `SERVER_PORT=8080` and `CL_SERVER_PORT=9000`, the value will be `8080`.

## Migration from Old Format

If you're using the old `CL_*` prefix format, you can:

1. **Keep using it** - Both formats are supported
2. **Migrate gradually** - Add new variables without prefix
3. **Mixed approach** - Use both formats simultaneously

Example migration:

```bash
# Old format (still works)
export CL_SERVER_PORT=3000
export CL_DATABASE_PATH=/data/db.sqlite

# New format (recommended)
export SERVER_PORT=3000
export DATABASE_PATH=/data/db.sqlite
```

## Complete Example

```bash
#!/bin/bash
# production-env.sh

# Server Configuration
export SERVER_PORT=8080
export SERVER_ENV=production

# Database
export DATABASE_PATH=/var/lib/central-logs/production.db

# Redis
export REDIS_URL=redis://:mypassword@redis.example.com:6379/0

# Security
export JWT_SECRET=$(openssl rand -base64 32)
export ADMIN_PASSWORD=$(openssl rand -base64 16)

# VAPID (generate at https://vapidkeys.com/)
export VAPID_PUBLIC_KEY="BCPvbPUl..."
export VAPID_PRIVATE_KEY="4dERHXcVs..."
export VAPID_SUBJECT="mailto:admin@example.com"

# Rate Limiting
export RATE_LIMIT_API_REQUESTS_PER_MINUTE=5000

# WebSocket
export WEBSOCKET_PING_INTERVAL=60s

# Start server
/opt/central-logs/bin/server
```
