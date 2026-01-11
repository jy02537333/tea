// 按需加载 marked，避免在首屏打包进主包
let _marked: any = null;
function getMarked() {
  if (!_marked) {
    // 同步加载以保持现有同步接口
    // Webpack 会将其打入独立 chunk（结合 splitChunks 配置）
    // eslint-disable-next-line @typescript-eslint/no-var-requires
    _marked = require('marked');
  }
  return _marked;
}
import { HEADING_COLOR, TEXT_COLOR, LINK_COLOR, H1_SIZE, H2_SIZE, H3_SIZE, H4_SIZE, PARAGRAPH_SIZE, LINE_HEIGHT, HEADING_MARGIN, PARAGRAPH_MARGIN, CODE_BG, CODE_BORDER, CODE_TEXT, CODE_FONT, TABLE_BORDER, TABLE_HEADER_BG, QUOTE_BG, QUOTE_BORDER, QUOTE_TEXT, INLINE_CODE_BG, INLINE_CODE_BORDER, INLINE_CODE_TEXT } from '../constants/theme';

export function markdownToHtml(md: string): string {
  try {
    const { marked, Renderer: MarkedRenderer } = getMarked();
    // Setup custom renderer to inject inline styles for Weapp RichText compatibility
    const renderer = new MarkedRenderer();
    renderer.heading = (text: string, level: number) => {
      const size = level === 1 ? H1_SIZE : level === 2 ? H2_SIZE : level === 3 ? H3_SIZE : H4_SIZE;
      return `<h${level} style="color:${HEADING_COLOR};font-size:${size}px;line-height:${LINE_HEIGHT};margin:${HEADING_MARGIN}">${text}</h${level}>`;
    };
    renderer.paragraph = (text: string) => {
      return `<p style="color:${TEXT_COLOR};font-size:${PARAGRAPH_SIZE}px;line-height:${LINE_HEIGHT};margin:${PARAGRAPH_MARGIN}">${text}</p>`;
    };
    renderer.link = (href: string | null, title: string | null, text: string) => {
      const t = title ? ` title="${title}"` : '';
      const safeHref = href || '#';
      return `<a href="${safeHref}"${t} style="color:${LINK_COLOR};text-decoration:none">${text}</a>`;
    };
    // Inline code
    renderer.codespan = (code: string) => {
      const escaped = escapeHtml(code);
      return `<code style="background:${INLINE_CODE_BG};border:1px solid ${INLINE_CODE_BORDER};border-radius:4px;padding:2px 6px;color:${INLINE_CODE_TEXT};font-family:${CODE_FONT}">${escaped}</code>`;
    };
    // Images with caption (use title or alt text); support local demo via data URL
    renderer.image = (href: string | null, title: string | null, text: string) => {
      let src = href || '';
      if (src.startsWith('local:')) {
        const key = src.split(':')[1] || 'demo';
        const labelMap: Record<string, string> = {
          'about-demo': 'About Demo',
          'privacy-demo': 'Privacy Demo',
          'help-demo': 'Help Demo',
          'terms-demo': 'Terms Demo',
          'demo': 'Demo'
        };
        const label = labelMap[key] || 'Demo';
        const svg = `<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"600\" height=\"320\"><rect width=\"100%\" height=\"100%\" fill=\"#e5e7eb\"/><text x=\"50%\" y=\"50%\" fill=\"#111827\" font-size=\"24\" text-anchor=\"middle\" dominant-baseline=\"middle\">${label}</text></svg>`;
        src = `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;
      }
      const caption = (title && title.trim()) || (text && text.trim()) || '';
      const img = `<img src="${src}" alt="${escapeHtml(text || '')}" style="max-width:100%;border-radius:8px" />`;
      if (!caption) return img;
      return `<figure style="margin:${PARAGRAPH_MARGIN};text-align:center">${img}<figcaption style="color:${TEXT_COLOR};font-size:${PARAGRAPH_SIZE - 1}px;margin-top:6px">${escapeHtml(caption)}</figcaption></figure>`;
    };
    // Blockquote
    renderer.blockquote = (quote: string) => {
      return `<blockquote style="background:${QUOTE_BG};border-left:4px solid ${QUOTE_BORDER};color:${QUOTE_TEXT};padding:8px 12px;margin:${PARAGRAPH_MARGIN};border-radius:6px">${quote}</blockquote>`;
    };
    // Code blocks
    renderer.code = (code: string, infostring?: string) => {
      const escaped = escapeHtml(code);
      return `<pre style="background:${CODE_BG};border:1px solid ${CODE_BORDER};border-radius:8px;padding:12px;overflow:auto;margin:${PARAGRAPH_MARGIN}"><code style="color:${CODE_TEXT};font-family:${CODE_FONT};font-size:${PARAGRAPH_SIZE - 1}px;line-height:${LINE_HEIGHT}">${escaped}</code></pre>`;
    };
    // Tables
    renderer.table = (header: string, body: string) => {
      return `<table style="width:100%;border-collapse:collapse;margin:${PARAGRAPH_MARGIN}"><thead>${header}</thead><tbody>${body}</tbody></table>`;
    };
    renderer.tablerow = (content: string) => {
      return `<tr>${content}</tr>`;
    };
    renderer.tablecell = (content: string, flags: { header: boolean; align?: 'center' | 'left' | 'right' | null }) => {
      const tag = flags.header ? 'th' : 'td';
      const align = flags.align || 'left';
      const bg = flags.header ? `background:${TABLE_HEADER_BG};` : '';
      return `<${tag} style="${bg}border:1px solid ${TABLE_BORDER};padding:8px 12px;text-align:${align};font-size:${PARAGRAPH_SIZE}px;line-height:${LINE_HEIGHT}">${content}</${tag}>`;
    };
    marked.use({ renderer });
    // Convert markdown to HTML
    return marked.parse(md || '') as string;
  } catch (_) {
    return (md || '').replace(/\n/g, '<br/>');
  }
}

function escapeHtml(s: string): string {
  return (s || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}
