# Redis Setup Guide

This guide explains how to configure Redis connection for different deployment scenarios.

## Quick Diagnosis

**Problem**: Redis connection failed with error like:
```
Warning: Failed to connect to Redis: dial tcp: lookup redis: no such host
```

**Cause**: Wrong Redis URL for your environment.

## Redis URL Configuration

### For Local Development

Use `redis://localhost:6379`

**Setup**:
1. Start Redis (if not running):
   ```bash
   docker run -d -p 6379:6379 --name redis redis:7-alpine
   # OR use docker-compose
   docker compose up -d redis
   ```

2. Set environment variable:
   ```bash
   export REDIS_URL=redis://localhost:6379
   ```

3. Or use `.env.local` file (recommended):
   ```bash
   # Copy from template
   cp .env.example .env.local

   # Edit .env.local
   REDIS_URL=redis://localhost:6379
   ```

4. Run the app:
   ```bash
   # Option 1: Export vars manually
   export $(cat .env.local | xargs)
   go run cmd/server/main.go

   # Option 2: Use make (recommended)
   make dev
   ```

### For Docker Compose Deployment

Use `redis://redis:6379` (service name)

**Setup**:
1. Create `.env` file:
   ```bash
   cp .env.production.example .env
   ```

2. Ensure `.env` has:
   ```bash
   REDIS_URL=redis://redis:6379
   ```

3. Start all services:
   ```bash
   docker compose up -d
   ```

### For Production (Standalone)

Use your Redis server's actual hostname/IP:

```bash
# External Redis server
export REDIS_URL=redis://your-redis-server.com:6379

# With password
export REDIS_URL=redis://:password@your-redis-server.com:6379

# Different database
export REDIS_URL=redis://your-redis-server.com:6379/1

# Redis Sentinel
export REDIS_URL=redis-sentinel://sentinel1:26379,sentinel2:26379/mymaster/0

# Redis Cluster
export REDIS_URL=redis://node1:6379?cluster=true
```

## Environment File Priority

The app supports multiple environment files for different scenarios:

1. **`.env.local`** - Local development (localhost Redis)
2. **`.env`** - Production deployment (Docker or remote Redis)
3. **`.env.production.example`** - Template for production

**Important**:
- `.env` and `.env.local` are in `.gitignore` (never commit these!)
- Only `.env.example` and `.env.production.example` should be committed

## Verification

Check if Redis is connected successfully:

```bash
# Look for these log messages when starting the server:

# ✅ Success (no Redis warnings)
Starting server on :3000 (env: production)

# ❌ Failed to connect
Warning: Failed to connect to Redis: dial tcp: lookup redis: no such host
Realtime features and rate limiting will be disabled
Starting server on :3000 (env: production)
```

If Redis fails to connect, the app will still start but with limited functionality:
- ❌ Rate limiting disabled
- ❌ Real-time features disabled
- ✅ Core API still works

## Testing Redis Connection

Test Redis connection manually:

```bash
# Test with redis-cli
redis-cli -h localhost -p 6379 ping
# Expected: PONG

# Test with docker
docker exec central-logs-redis redis-cli ping
# Expected: PONG

# Test from Go code
go run -e 'package main; import "github.com/redis/go-redis/v9"; ...'
```

## Common Issues

### Issue 1: "no such host" error

**Error**:
```
dial tcp: lookup redis: no such host
```

**Cause**: Using `redis://redis:6379` outside Docker network

**Solution**: Use `redis://localhost:6379` for local development

### Issue 2: "connection refused"

**Error**:
```
dial tcp 127.0.0.1:6379: connect: connection refused
```

**Cause**: Redis not running

**Solution**:
```bash
docker compose up -d redis
# OR
docker run -d -p 6379:6379 redis:7-alpine
```

### Issue 3: "authentication required"

**Error**:
```
NOAUTH Authentication required
```

**Cause**: Redis has password but not provided in URL

**Solution**:
```bash
export REDIS_URL=redis://:your-password@localhost:6379
```

### Issue 4: Port already in use

**Error**:
```
Error starting userland proxy: listen tcp4 0.0.0.0:6379: bind: address already in use
```

**Cause**: Another Redis instance running on port 6379

**Solution**:
```bash
# Find and stop the conflicting process
lsof -i :6379
kill <PID>

# Or use a different port
docker run -d -p 6380:6379 redis:7-alpine
export REDIS_URL=redis://localhost:6380
```

## Redis Configuration

For production, you may want to configure Redis persistence and other options:

```yaml
# docker-compose.yml
services:
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --requirepass yourpassword
    volumes:
      - redis_data:/data
```

Then update `.env`:
```bash
REDIS_URL=redis://:yourpassword@redis:6379
```

## Monitoring Redis

Check Redis stats:

```bash
# Connect to Redis
docker exec -it central-logs-redis redis-cli

# Get info
INFO

# Monitor commands
MONITOR

# Check memory usage
INFO memory
```

## Alternative: Disable Redis (Not Recommended)

If you don't need rate limiting or real-time features, you can run without Redis:

1. Don't set `REDIS_URL` environment variable
2. Or set it to empty: `export REDIS_URL=""`

The app will start with a warning:
```
Warning: Failed to connect to Redis: ...
Realtime features and rate limiting will be disabled
```

**Note**: This is NOT recommended for production as it disables important security features.

## Summary

| Environment | Redis URL | Notes |
|-------------|-----------|-------|
| Local Dev | `redis://localhost:6379` | Requires local Redis |
| Docker Compose | `redis://redis:6379` | Uses service name |
| Production | `redis://your-server:6379` | Use actual hostname |
| With Password | `redis://:pass@host:6379` | Include password |
| Different DB | `redis://host:6379/1` | Specify database number |

For most use cases:
- **Development**: Use `.env.local` with `redis://localhost:6379`
- **Production**: Use `.env` with `redis://redis:6379` (Docker) or your Redis server URL
