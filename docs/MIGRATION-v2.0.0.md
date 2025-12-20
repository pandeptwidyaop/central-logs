# Migration Guide: v1.x to v2.0.0

## Overview

Version 2.0.0 introduces **breaking changes** to WebSocket authentication for improved security. This guide helps you migrate from v1.x to v2.0.0.

## Breaking Changes

### 1. WebSocket Authentication Method Changed

**What Changed:**
- WebSocket authentication moved from query parameters to `Sec-WebSocket-Protocol` header
- Removed `user_id` query parameter (security improvement - no longer exposes user IDs in URLs/logs)

**Why This Change:**
- **Security**: Query parameters appear in server logs, browser history, and referrer headers
- **Best Practice**: Using WebSocket subprotocols is the standard way to pass authentication tokens
- **Compliance**: Aligns with WebSocket RFC 6455 specifications

---

## Migration Steps

### For Backend Developers

#### No Changes Required ✅

If you're using the official Central Logs backend:
- The backend automatically handles both old and new authentication methods during the transition period
- Simply update to v2.0.0 - no code changes needed

#### If You Built Custom WebSocket Clients

Update your WebSocket connection code:

**Before (v1.x):**
```javascript
// ❌ OLD - Don't use this
const ws = new WebSocket(`ws://localhost:3000/ws/logs?user_id=${userId}&token=${token}`);
```

**After (v2.0.0):**
```javascript
// ✅ NEW - Use Sec-WebSocket-Protocol
const ws = new WebSocket('ws://localhost:3000/ws/logs', ['token', jwtToken]);

// Optional: Filter by project
const ws = new WebSocket('ws://localhost:3000/ws/logs?project_id=${projectId}', ['token', jwtToken]);
```

---

### For Frontend Developers

#### If Using Official Frontend

Update to the latest frontend build:

```bash
cd frontend
npm install
npm run build
```

Then redeploy your frontend.

#### If You Built Custom Frontend

Update your WebSocket hook/component:

**Before (v1.x):**
```typescript
// ❌ OLD
const { user } = useAuth();
const wsUrl = `${protocol}//${host}/ws/logs?user_id=${user.id}`;
const ws = new WebSocket(wsUrl);
```

**After (v2.0.0):**
```typescript
// ✅ NEW
const { user, token } = useAuth(); // Get token from auth context
const wsUrl = `${protocol}//${host}/ws/logs`; // No user_id in URL

// Add project filter if needed
if (projectId) {
  wsUrl += `?project_id=${projectId}`;
}

// Pass token via Sec-WebSocket-Protocol
const ws = new WebSocket(wsUrl, ['token', token]);
```

**Update Your Auth Context:**

Ensure your auth context exposes the JWT token:

```typescript
interface AuthContextType {
  user: User | null;
  token: string | null; // ← Add this
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
}

// In your provider:
const [token, setToken] = useState<string | null>(null);

// Update on login:
const login = async (username, password) => {
  const response = await api.login(username, password);
  setToken(api.getToken()); // Store token
};

// Clear on logout:
const logout = () => {
  api.logout();
  setToken(null);
};

// Provide token in context:
return (
  <AuthContext.Provider value={{ user, token, login, logout }}>
    {children}
  </AuthContext.Provider>
);
```

---

## Testing Your Migration

### 1. Check WebSocket Connection

Open browser DevTools → Network tab → WS filter:

**Success Indicators:**
- Status: `101 Switching Protocols`
- Request Headers include: `Sec-WebSocket-Protocol: token, <your-jwt>`
- Response Headers include: `Sec-WebSocket-Protocol: token`

**Failure Indicators:**
- Status: `401 Unauthorized` - Check your JWT token is valid
- Connection closes immediately - Ensure backend is v2.0.0 and includes the protocol response header
- `403 Forbidden` - User may be inactive or token expired

### 2. Verify Realtime Logs

1. Open the Logs page
2. Check for green "Live" indicator (top right)
3. Generate test logs using loggen or API:

```bash
# Using loggen
./bin/loggen -url http://localhost:3000/api/v1/logs \
             -key YOUR_API_KEY \
             -duration 30s

# Or using curl
curl -X POST http://localhost:3000/api/v1/logs \
     -H "X-API-Key: YOUR_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{
       "level": "INFO",
       "message": "Test log from migration",
       "metadata": {"test": true}
     }'
```

4. Logs should appear instantly without page refresh

### 3. Check Browser Console

Look for any WebSocket errors:

```javascript
// No errors = Success ✅
// "WebSocket connection failed" = Check authentication
// "Unexpected EOF" = Backend needs Sec-WebSocket-Protocol response header fix
```

---

## Rollback Plan

If you need to rollback to v1.x:

### Backend
```bash
git checkout v1.5.0  # Or your last v1.x version
make build
./bin/central-logs
```

### Frontend
```bash
cd frontend
git checkout v1.5.0
npm install
npm run build
# Copy frontend/dist to web/dist
```

---

## Security Improvements

Version 2.0.0 provides these security enhancements:

1. **No Token Exposure in Logs**
   - v1.x: `GET /ws/logs?user_id=123&token=xxx` (token in access logs)
   - v2.0.0: `GET /ws/logs` with token in headers (not logged)

2. **No User ID Exposure**
   - v1.x: User IDs visible in browser URLs and server logs
   - v2.0.0: User identity only in encrypted JWT token

3. **Standard Compliance**
   - v2.0.0 follows RFC 6455 WebSocket subprotocol specification
   - Better compatibility with proxies and load balancers

4. **Constant-Time Token Comparison**
   - All token comparisons use `crypto/subtle` to prevent timing attacks
   - Applies to both API keys and MCP tokens

5. **Security Headers**
   - Added comprehensive security headers (CSP, X-Frame-Options, etc.)
   - See `docs/security-improvements.md` for details

---

## Need Help?

If you encounter issues during migration:

1. Check the [Troubleshooting](#troubleshooting) section below
2. Review server logs for WebSocket connection errors
3. Open an issue: https://github.com/pandeptwidyaop/central-logs/issues

---

## Troubleshooting

### WebSocket shows "Connecting..." forever

**Cause:** Backend not responding with `Sec-WebSocket-Protocol` header

**Solution:**
- Ensure backend is v2.0.0 or later
- Check `/internal/websocket/handler.go` includes:
  ```go
  if c.Get("Sec-WebSocket-Protocol") != "" {
      c.Set("Sec-WebSocket-Protocol", "token")
  }
  ```

### 401 Unauthorized on WebSocket

**Cause:** Token not being sent or is invalid

**Solution:**
- Verify token is exposed in auth context: `const { token } = useAuth()`
- Check token is not null/undefined
- Ensure WebSocket creation uses: `new WebSocket(url, ['token', token])`
- Verify JWT hasn't expired (default: 24 hours)

### Logs not appearing in realtime

**Cause:** WebSocket not connected or project filter mismatch

**Solution:**
- Check green "Live" indicator is showing
- Verify project filter matches logs being generated
- Check browser console for errors
- Ensure Redis is running (optional, but needed for realtime features)

### Duplicate log messages

**Cause:** Multiple WebSocket connections or old connections not cleaned up

**Solution:**
- Clear browser cache and reload
- Check only one WebSocket connection in DevTools → Network → WS
- Verify connection cleanup in `useEffect` dependencies

---

## What's Next?

After migrating to v2.0.0, consider:

1. **Update API Documentation** - If you have custom API docs, update WebSocket examples
2. **Test E2E** - Run full end-to-end tests with Playwright/Cypress
3. **Monitor Logs** - Check server logs for any connection issues in production
4. **Update Client Libraries** - If you have client SDKs, update them to use new auth method

---

## Version History

- **v2.0.0** (December 21, 2025) - Breaking: WebSocket authentication via Sec-WebSocket-Protocol
- **v1.5.0** (December 20, 2025) - Security improvements, MCP support, 2FA
- **v1.4.0** - Push notifications, realtime logs
- **v1.0.0** - Initial release

---

**Migration Completed?** 🎉

Delete this file after successful migration, or keep it for reference.
