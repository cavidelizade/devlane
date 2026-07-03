export function formatRelativeTime(iso: string): string {
  const d = new Date(iso);
  const now = new Date();
  const s = Math.floor((now.getTime() - d.getTime()) / 1000);
  if (s < 60) return 'just now';
  if (s < 3600) {
    const minutes = Math.floor(s / 60);
    return `${minutes} ${minutes === 1 ? 'minute' : 'minutes'} ago`;
  }
  if (s < 86400) {
    const hours = Math.floor(s / 3600);
    return `${hours} ${hours === 1 ? 'hour' : 'hours'} ago`;
  }
  if (s < 2592000) {
    const days = Math.floor(s / 86400);
    return `${days} ${days === 1 ? 'day' : 'days'} ago`;
  }
  if (s < 31536000) {
    const months = Math.floor(s / 2592000);
    return `${months} ${months === 1 ? 'month' : 'months'} ago`;
  }
  const years = Math.floor(s / 31536000);
  return `${years} ${years === 1 ? 'year' : 'years'} ago`;
}

// Build timezone options: UTC offset + label (e.g. "UTC-07:00 America/Los_Angeles")
export function getTimezoneOptions(): { value: string; label: string }[] {
  try {
    const ids =
      typeof Intl !== 'undefined' &&
      typeof (Intl as { supportedValuesOf?: (key: string) => string[] }).supportedValuesOf ===
        'function'
        ? (Intl as { supportedValuesOf: (key: string) => string[] }).supportedValuesOf('timeZone')
        : [
            'UTC',
            'America/New_York',
            'America/Chicago',
            'America/Denver',
            'America/Los_Angeles',
            'Europe/London',
            'Europe/Paris',
            'Asia/Tokyo',
            'Asia/Kolkata',
            'Australia/Sydney',
          ];
    const now = new Date();
    return ids
      .map((value) => {
        try {
          const formatter = new Intl.DateTimeFormat('en-US', {
            timeZone: value,
            timeZoneName: 'longOffset',
          });
          const parts = formatter.formatToParts(now);
          const offsetPart = parts.find((p) => p.type === 'timeZoneName');
          const offset = offsetPart?.value ?? 'UTC';
          return { value, label: `${offset} ${value}` };
        } catch {
          return { value, label: value };
        }
      })
      .sort((a, b) => a.label.localeCompare(b.label));
  } catch {
    return [
      { value: 'UTC', label: 'UTC UTC' },
      { value: 'America/New_York', label: 'America/New York' },
    ];
  }
}
