/*
 * The route to the final: one long dashed path from a coral start dot to a
 * small pennant, with a single mustard waypoint. A minimal oversized
 * gesture used where dates and progression are the subject.
 */
export function RouteScene({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 480 200" aria-hidden="true" focusable="false">
      <path
        d="M 24 152 C 130 152 150 58 250 58 C 340 58 360 142 456 110"
        fill="none"
        stroke="var(--color-ink)"
        strokeWidth="5"
        strokeLinecap="round"
        strokeDasharray="2 14"
      />
      <circle cx="24" cy="152" r="14" fill="var(--color-accent)" />
      <circle cx="250" cy="58" r="9" fill="var(--color-mustard)" />
      <rect x="444" y="48" width="8" height="70" rx="4" fill="var(--color-ink)" />
      <path d="M 452 54 L 476 66 L 452 78 Z" fill="var(--color-accent)" />
    </svg>
  );
}
