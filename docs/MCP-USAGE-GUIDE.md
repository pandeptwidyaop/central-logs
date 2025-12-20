# MCP Integration Guide - Using Central Logs with AI Agents

## Apa itu MCP?

MCP (Model Context Protocol) adalah protokol yang memungkinkan AI agents (seperti Claude Desktop, custom agents, dll) untuk berkomunikasi dengan aplikasi eksternal. Dengan MCP, AI agent bisa:

- **Query logs** dari Central Logs
- **Search** berdasarkan kriteria tertentu
- **Filter** berdasarkan project, level, waktu, dll
- **Analisis** pattern dalam logs

## Prerequisites

1. Central Logs server running (default: `http://localhost:3000`)
2. MCP feature enabled di admin panel
3. MCP Token yang sudah dibuat

---

## Step 1: Enable MCP Server

### Via Admin Panel (Web UI)

1. Login ke Central Logs sebagai Admin
2. Buka **Settings** → **MCP Integration**
3. Klik **Enable MCP Server**
4. Status akan berubah menjadi "Enabled"

### Via API

```bash
curl -X POST http://localhost:3000/api/admin/mcp/toggle \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"
```

---

## Step 2: Create MCP Token

### Via Admin Panel

1. Buka **Settings** → **MCP Integration**
2. Klik **Create Token**
3. Isi:
   - **Name**: e.g., "Claude Desktop", "AI Agent 1"
   - **Description**: e.g., "Token untuk Claude Desktop di MacBook"
4. Klik **Create**
5. **PENTING**: Copy token yang muncul! Token hanya ditampilkan SEKALI saat pembuatan

### Via API

```bash
curl -X POST http://localhost:3000/api/admin/mcp/tokens \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Claude Desktop",
    "description": "Token for my AI agent"
  }'
```

Response:
```json
{
  "id": "token-uuid",
  "name": "Claude Desktop",
  "token": "mcp_abc123...",  ← COPY THIS!
  "created_at": "2025-12-21T..."
}
```

---

## Step 3: Configure AI Agent (Claude Desktop Example)

### For Claude Desktop

Edit file konfigurasi Claude Desktop:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

Add Central Logs MCP server:

```json
{
  "mcpServers": {
    "central-logs": {
      "url": "http://localhost:3000/api/mcp/message",
      "headers": {
        "Authorization": "Bearer mcp_abc123..."
      }
    }
  }
}
```

**IMPORTANT**: Ganti `mcp_abc123...` dengan token yang kamu copy dari Step 2!

### Restart Claude Desktop

Setelah save config, restart Claude Desktop agar MCP server ter-load.

---

## Step 4: Verify Connection

Di Claude Desktop, coba tanya:

```
Show me tools available
```

atau

```
What MCP servers are connected?
```

Claude harus menampilkan **central-logs** dengan tool **query_logs**.

---

## Step 5: Using MCP Tools

### Available Tools

#### 1. `query_logs` - Search and Filter Logs

**Parameters**:
- `project_ids` (array, optional): Filter by project IDs
- `levels` (array, optional): Filter by levels (`debug`, `info`, `warn`, `error`)
- `source` (string, optional): Filter by log source
- `search` (string, optional): Full-text search in message/metadata
- `start_time` (string, optional): Start time (RFC3339 format)
- `end_time` (string, optional): End time (RFC3339 format)
- `limit` (number, optional): Max results (default: 100, max: 1000)
- `offset` (number, optional): Pagination offset

**Example Queries for Claude**:

```
Show me the last 10 error logs
```

```
Find all logs from project "9d065a54-7c72-4406-afc1-877180327c8c" in the last hour
```

```
Search for logs containing "database timeout" with ERROR level
```

```
Show me all CRITICAL and ERROR logs from yesterday
```

```
Find logs with message containing "payment failed"
```

---

## Usage Examples

### Example 1: Debugging Errors

**User**: "I'm getting errors in my API. Can you check the logs?"

**Claude** (using MCP):
```
Let me query your logs for recent errors...

*calls query_logs with level=["error"], limit=20*

I found 5 ERROR logs in the last hour:
1. [1:30 PM] "Request validation failed: 686 errors"
2. [1:25 PM] "Database connection timeout"
...
```

### Example 2: Finding Patterns

**User**: "Are there any patterns in failed login attempts?"

**Claude**:
```
*calls query_logs with search="login failed", limit=100*

Yes, I found 15 failed login attempts in the last 24 hours.
Most common pattern:
- Peak time: 2-3 AM (12 attempts)
- All from same IP: 192.168.1.x
- Suggests possible brute force attack
```

### Example 3: Project-Specific Monitoring

**User**: "What's happening with Test Project?"

**Claude**:
```
*calls query_logs with project_ids=["9d065a54-..."], limit=50*

Latest activity in Test Project:
- 23 INFO logs in last hour
- 2 WARN logs about high memory usage
- No ERROR logs
Overall: Healthy ✅
```

---

## Advanced Configuration

### Multiple MCP Tokens

You can create multiple tokens for different purposes:

```json
{
  "mcpServers": {
    "central-logs-prod": {
      "url": "https://logs.example.com/api/mcp/message",
      "headers": {
        "Authorization": "Bearer mcp_prod_token..."
      }
    },
    "central-logs-staging": {
      "url": "https://staging.logs.example.com/api/mcp/message",
      "headers": {
        "Authorization": "Bearer mcp_staging_token..."
      }
    }
  }
}
```

### Token Management

#### View All Tokens

```bash
curl http://localhost:3000/api/admin/mcp/tokens \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Revoke Token

```bash
curl -X DELETE http://localhost:3000/api/admin/mcp/tokens/TOKEN_ID \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### View Token Activity

```bash
curl http://localhost:3000/api/admin/mcp/tokens/TOKEN_ID/activity \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Security Best Practices

1. **Rotate Tokens Regularly**
   - Create new token
   - Update AI agent config
   - Revoke old token

2. **Use Descriptive Names**
   - Good: "Claude Desktop - MacBook Pro"
   - Bad: "Token 1"

3. **Monitor Token Activity**
   - Check activity logs regularly
   - Look for suspicious patterns

4. **Revoke Unused Tokens**
   - Delete tokens that are no longer needed
   - Prevents unauthorized access

5. **Use HTTPS in Production**
   - Configure: `https://your-domain.com/api/mcp/message`
   - Not: `http://...` (insecure)

---

## Troubleshooting

### "MCP server not responding"

**Cause**: Server tidak running atau MCP disabled

**Solution**:
1. Check server is running: `curl http://localhost:3000/api/version`
2. Check MCP enabled: Visit admin panel → Settings → MCP
3. Check token valid

### "Unauthorized" Error

**Cause**: Token salah atau expired

**Solution**:
1. Verify token di config matches token dari admin panel
2. Check token tidak di-revoke
3. Format header: `Authorization: Bearer mcp_...` (ada space setelah Bearer!)

### "Tool not found"

**Cause**: AI agent belum reload MCP servers

**Solution**:
1. Restart AI agent (Claude Desktop, etc.)
2. Verify MCP config file saved correctly
3. Check JSON syntax valid

### No Logs Returned

**Cause**: Query filter terlalu strict atau project kosong

**Solution**:
1. Try query without filters first: "show me latest logs"
2. Verify project has logs: check via web UI
3. Check time range not too narrow

---

## Performance Tips

1. **Use Pagination**
   - Query dengan `limit` kecil (10-100)
   - Use `offset` untuk next page

2. **Filter by Project**
   - Specify `project_ids` untuk hasil lebih focused
   - Faster query

3. **Use Time Ranges**
   - Specify `start_time` dan `end_time`
   - Reduces data scanned

4. **Cache Results**
   - AI agent akan cache hasil query
   - Subsequent questions lebih cepat

---

## Custom AI Agents

### Using with Python

```python
import requests

MCP_ENDPOINT = "http://localhost:3000/api/mcp/message"
MCP_TOKEN = "mcp_your_token_here"

def query_logs(search=None, levels=None, limit=100):
    headers = {
        "Authorization": f"Bearer {MCP_TOKEN}",
        "Content-Type": "application/json"
    }

    payload = {
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": "query_logs",
            "arguments": {
                "search": search,
                "levels": levels or [],
                "limit": limit
            }
        }
    }

    response = requests.post(MCP_ENDPOINT, json=payload, headers=headers)
    return response.json()

# Example usage
results = query_logs(search="error", levels=["error", "critical"], limit=50)
print(results)
```

### Using with Node.js

```javascript
const axios = require('axios');

const MCP_ENDPOINT = 'http://localhost:3000/api/mcp/message';
const MCP_TOKEN = 'mcp_your_token_here';

async function queryLogs({ search, levels, limit = 100 }) {
  const response = await axios.post(
    MCP_ENDPOINT,
    {
      jsonrpc: '2.0',
      id: 1,
      method: 'tools/call',
      params: {
        name: 'query_logs',
        arguments: {
          search,
          levels: levels || [],
          limit
        }
      }
    },
    {
      headers: {
        'Authorization': `Bearer ${MCP_TOKEN}`,
        'Content-Type': 'application/json'
      }
    }
  );

  return response.data;
}

// Example usage
queryLogs({
  search: 'database',
  levels: ['error'],
  limit: 20
}).then(console.log);
```

---

## FAQ

**Q: Apakah MCP token sama dengan API key?**
A: Tidak. API key untuk log ingestion, MCP token untuk AI agents querying logs.

**Q: Berapa banyak token yang bisa saya buat?**
A: Unlimited. Tapi best practice: 1 token per agent/device.

**Q: Apakah token expired?**
A: Tidak otomatis. Token active sampai di-revoke manually.

**Q: Bisa pakai MCP dengan Cursor/VS Code?**
A: Ya! Jika IDE support MCP protocol, tinggal configure endpoint + token.

**Q: Performance impact kalau banyak AI agents?**
A: Minimal. Each query di-rate limit dan di-log untuk monitoring.

---

## Next Steps

1. ✅ Enable MCP server
2. ✅ Create your first token
3. ✅ Configure AI agent (Claude Desktop)
4. ✅ Test with simple queries
5. 🚀 Build automation workflows!

---

**Need Help?**

- Check activity logs: Settings → MCP Integration → Token Activity
- Server logs: `tail -f logs/central-logs.log`
- Issues: https://github.com/pandeptwidyaop/central-logs/issues

---

**Selamat menggunakan MCP! 🎉**

AI agents kamu sekarang bisa query, analyze, dan monitor logs secara real-time!
