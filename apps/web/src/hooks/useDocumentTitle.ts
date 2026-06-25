import { useEffect } from 'react';

/** App name shown as the browser tab title suffix. */
export const APP_NAME = 'Devlane';

/** Builds a tab title of the form `"<title> — Devlane"`, or just the app name. */
export function formatDocumentTitle(title?: string | null): string {
  const trimmed = title?.trim();
  return trimmed ? `${trimmed} — ${APP_NAME}` : APP_NAME;
}

/**
 * Sets `document.title` to `"<title> — Devlane"` while the calling page is
 * mounted, falling back to just "Devlane" when no title is given. The title is
 * recomputed whenever `title` changes, so detail pages can pass a value that
 * arrives asynchronously (e.g. once the work item or project has loaded).
 */
export function useDocumentTitle(title?: string | null): void {
  useEffect(() => {
    document.title = formatDocumentTitle(title);
  }, [title]);
}
