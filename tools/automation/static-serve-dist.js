// Minimal static file server for wx-fe/dist
// Usage: node tools/automation/static-serve-dist.js [port]

const http = require('http');
const fs = require('fs');
const path = require('path');

const port = Number(process.argv[2] || 10090);
const root = path.resolve(__dirname, '../../wx-fe/dist');

const mime = {
  '.html': 'text/html',
  '.js': 'application/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.gif': 'image/gif',
  '.svg': 'image/svg+xml',
  '.ico': 'image/x-icon',
};

const server = http.createServer((req, res) => {
  const url = decodeURIComponent(req.url || '/');
  let filePath = path.join(root, url.split('?')[0].replace(/^\//, ''));
  if (url === '/' || !path.extname(filePath)) {
    filePath = path.join(root, 'index.html');
  }
  fs.readFile(filePath, (err, data) => {
    if (err) {
      // SPA fallback
      fs.readFile(path.join(root, 'index.html'), (e2, d2) => {
        if (e2) {
          res.writeHead(404);
          res.end('Not found');
          return;
        }
        res.writeHead(200, { 'Content-Type': 'text/html' });
        res.end(d2);
      });
      return;
    }
    const ext = path.extname(filePath).toLowerCase();
    res.writeHead(200, { 'Content-Type': mime[ext] || 'application/octet-stream', 'Cache-Control': 'no-store' });
    res.end(data);
  });
});

server.listen(port, () => {
  console.log(`[static] serving ${root} at http://127.0.0.1:${port}`);
});
