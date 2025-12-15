const http = require('http');
const fs = require('fs');
const path = require('path');

const PORT = 7242;
const LOG_FILE = path.join(__dirname, 'debug.log');

// Clear log file on start
fs.writeFileSync(LOG_FILE, `=== Debug Server Started at ${new Date().toISOString()} ===\n\n`);

const server = http.createServer((req, res) => {
  // CORS headers
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'POST, GET, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');

  if (req.method === 'OPTIONS') {
    res.writeHead(204);
    res.end();
    return;
  }

  if (req.method === 'POST' && req.url.startsWith('/ingest')) {
    let body = '';
    req.on('data', chunk => body += chunk);
    req.on('end', () => {
      try {
        const data = JSON.parse(body);
        const timestamp = new Date().toISOString();
        const logEntry = `[${timestamp}] ${data.location || 'unknown'}\n` +
          `  Message: ${data.message || 'no message'}\n` +
          `  Data: ${JSON.stringify(data.data || {}, null, 2)}\n` +
          `  Stack: ${data.stack || 'none'}\n` +
          `---\n`;

        fs.appendFileSync(LOG_FILE, logEntry);
        console.log(`[LOG] ${data.location}: ${data.message}`);

        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ ok: true }));
      } catch (err) {
        console.error('Parse error:', err);
        res.writeHead(400);
        res.end(JSON.stringify({ error: 'Invalid JSON' }));
      }
    });
    return;
  }

  if (req.method === 'GET' && req.url === '/logs') {
    try {
      const logs = fs.readFileSync(LOG_FILE, 'utf8');
      res.writeHead(200, { 'Content-Type': 'text/plain' });
      res.end(logs);
    } catch (err) {
      res.writeHead(500);
      res.end('Error reading logs');
    }
    return;
  }

  if (req.method === 'GET' && req.url === '/clear') {
    fs.writeFileSync(LOG_FILE, `=== Logs Cleared at ${new Date().toISOString()} ===\n\n`);
    res.writeHead(200);
    res.end('Logs cleared');
    return;
  }

  res.writeHead(404);
  res.end('Not found');
});

server.listen(PORT, () => {
  console.log(`Debug server running on http://localhost:${PORT}`);
  console.log(`  POST /ingest/{id} - Send logs`);
  console.log(`  GET /logs - View logs`);
  console.log(`  GET /clear - Clear logs`);
  console.log(`  Log file: ${LOG_FILE}`);
});
