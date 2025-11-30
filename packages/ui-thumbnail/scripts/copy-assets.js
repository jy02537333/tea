const fs = require('fs');
const path = require('path');

function copyFile(src, dest) {
  const dir = path.dirname(dest);
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
  fs.copyFileSync(src, dest);
  console.log('copied', src, '->', dest);
}

const root = path.resolve(__dirname, '..');
const reactCss = path.join(root, 'react', 'thumbnail.css');
const outReactCss = path.join(root, 'dist', 'react', 'thumbnail.css');

if (fs.existsSync(reactCss)) copyFile(reactCss, outReactCss);

console.log('assets copied');
