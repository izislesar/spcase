/*
 * Brand mark: a small championship pennant — the actual brand asset, used
 * only as the navigation logo. The pole follows the surrounding text color
 * so the mark works on every surface; the flag takes the accent.
 */
export function Mark({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true" focusable="false">
      <rect x="4.5" y="2.5" width="3" height="19" rx="1.5" fill="currentColor" />
      <path d="M 9 4 L 21.5 8.75 L 9 13.5 Z" fill="var(--color-accent)" />
    </svg>
  );
}
