/*
 * Brand mark: a small championship pennant — the simplest product symbol
 * (reaching the final), not an abstract identity shape. The pole takes
 * --mark-pole so the mark works on light and navy surfaces.
 */
export function Mark({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true" focusable="false">
      <rect
        x="4.5"
        y="2.5"
        width="3"
        height="19"
        rx="1.5"
        fill="var(--mark-pole, var(--color-ink))"
      />
      <path d="M 9 4 L 21.5 8.75 L 9 13.5 Z" fill="var(--color-accent)" />
    </svg>
  );
}
