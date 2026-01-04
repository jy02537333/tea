import Taro from '@tarojs/taro';

let dompurify: any = null;
try {
  // Lazy import; only available on H5
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  dompurify = require('dompurify');
} catch (_) {}

export function sanitizeHtml(html: string): string {
  const env = (Taro as any)?.getEnv ? (Taro as any).getEnv() : 'WEB';
  if (env === 'WEB' && dompurify) {
    const DOMPurify = dompurify.default || dompurify;
    return DOMPurify.sanitize(html || '', {
      USE_PROFILES: { html: true },
      FORBID_TAGS: ['script', 'iframe', 'object', 'embed', 'style'],
      FORBID_ATTR: ['onerror', 'onload', 'onclick', 'onmouseover', 'onfocus', 'onblur'],
      ALLOWED_URI_REGEXP: /^(?:(?:https?|mailto|tel):|[^a-z]|[a-z+.-]+(?:[^a-z+.-]|$))/i,
    });
  }

  // Fallback: strip dangerous tags/attributes
  let safe = (html || '')
    // remove scripts/styles/iframes/objects
    .replace(/<\s*(script|style|iframe|object|embed)[^>]*>[^<]*<\s*\/\s*\1\s*>/gi, '')
    // remove inline event handlers
    .replace(/\son[a-z]+\s*=\s*"[^"]*"/gi, '')
    .replace(/\son[a-z]+\s*=\s*'[^']*'/gi, '')
    .replace(/\son[a-z]+\s*=\s*[^\s>]+/gi, '')
    // neutralize javascript: URLs
    .replace(/href\s*=\s*"javascript:[^"]*"/gi, 'href="#"')
    .replace(/href\s*=\s*'javascript:[^']*'/gi, "href='#'")
    .replace(/href\s*=\s*javascript:[^\s>]*/gi, 'href=#');
  return safe;
}
