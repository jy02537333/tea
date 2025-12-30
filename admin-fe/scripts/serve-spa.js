// tiny static SPA server with history fallback for dist/
// usage: node scripts/serve-spa.js [port]
import http from 'http';
import fs from 'fs';
import path from 'path';
import url from 'url';

const __dirname = path.dirname(url.fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, '..', 'dist');
const port = Number(process.argv[2] || process.env.PORT || 5173);

const mime = (ext) => ({
  '.html': 'text/html; charset=utf-8',
  '.js': 'application/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.ico': 'image/x-icon',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.svg': 'image/svg+xml',
  '.map': 'application/json; charset=utf-8',
}[ext] || 'application/octet-stream');

const indexHtml = fs.readFileSync(path.join(root, 'index.html'));

const server = http.createServer((req, res) => {
  const parsed = url.parse(req.url || '/');
  let pathname = decodeURIComponent(parsed.pathname || '/');
  if (pathname === '/') pathname = '/index.html';

  const filePath = path.join(root, pathname);
  if (filePath.startsWith(root) && fs.existsSync(filePath) && fs.statSync(filePath).isFile()) {
    const ext = path.extname(filePath);
    res.statusCode = 200;
    res.setHeader('Content-Type', mime(ext));
    fs.createReadStream(filePath).pipe(res);
  } else {
    // SPA history fallback
    res.statusCode = 200;
    res.setHeader('Content-Type', 'text/html; charset=utf-8');
    res.end(indexHtml);
  }
});

server.listen(port, '127.0.0.1', () => {
  // eslint-disable-next-line no-console
  console.log(`[serve-spa] listening on http://127.0.0.1:${port}`);
});
