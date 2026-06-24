import DOMPurify from 'dompurify';

/**
 * Sanitize untrusted HTML (comment bodies, page/version content, etc.) before
 * injecting it via dangerouslySetInnerHTML. Strips <script>, inline event
 * handlers, and javascript:/data: URLs while preserving normal rich-text markup.
 */
export function sanitizeHtml(html: string | null | undefined): string {
  if (!html) return '';
  return DOMPurify.sanitize(html, { USE_PROFILES: { html: true } });
}

/**
 * Returns the URL only if it uses a safe scheme (http/https/mailto) or is a
 * site-relative path; otherwise returns '#'. Prevents `javascript:`/`data:` URI
 * injection through user-supplied link hrefs.
 */
export function safeUrl(url: string | null | undefined): string {
  if (!url) return '#';
  const trimmed = url.trim();
  // Site-relative path (but not protocol-relative "//host").
  if (trimmed.startsWith('/') && !trimmed.startsWith('//')) return trimmed;
  try {
    const parsed = new URL(trimmed, window.location.origin);
    if (
      parsed.protocol === 'http:' ||
      parsed.protocol === 'https:' ||
      parsed.protocol === 'mailto:'
    ) {
      return trimmed;
    }
  } catch {
    // unparseable — fall through to the safe default
  }
  return '#';
}
