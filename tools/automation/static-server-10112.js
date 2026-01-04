const http = require('http');
const fs = require('fs');
const path = require('path');
const root = path.resolve(__dirname, '../../wx-fe/dist');
const port = 10112;
http.createServer((req, res) => {
  const reqPath = req.url.split('?')[0];
  let p = path.join(root, reqPath);
  const sendFile = (fp) => fs.readFile(fp, (err, data) => {
    if (err) { res.writeHead(404); res.end('not found'); return; }
    res.writeHead(200); res.end(data);
  });
  if (reqPath === '/' || reqPath === '') {
    return sendFile(path.join(root, 'index.html'));
  }
  fs.stat(p, (err, stats) => {
    if (!err && stats.isFile()) return sendFile(p);
    // SPA fallback for browser history routes
    return sendFile(path.join(root, 'index.html'));
  });
}).listen(port, '127.0.0.1', () => console.log('static server on', port));
