/*
 * Closing scene: the finish composition for the navy footer — a giant
 * mustard field rising behind the podium, the pennant planted on the
 * center step, flat confetti on the open left side. Very wide, cropped at
 * the bottom edge; built to sit behind closing typography.
 */
export function ClosingScene({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 900 420" aria-hidden="true" focusable="false">
      <circle cx="630" cy="450" r="280" fill="var(--color-mustard)" />
      <ellipse cx="620" cy="392" rx="220" ry="14" fill="var(--color-ink)" fillOpacity="0.25" />
      {/* confetti over the open left half */}
      <rect
        x="90"
        y="120"
        width="20"
        height="20"
        fill="var(--color-surface)"
        transform="rotate(18 100 130)"
      />
      <rect
        x="150"
        y="70"
        width="16"
        height="16"
        fill="var(--color-mustard)"
        transform="rotate(-28 158 78)"
      />
      <rect
        x="210"
        y="150"
        width="18"
        height="18"
        fill="var(--color-turquoise)"
        transform="rotate(42 219 159)"
      />
      <rect
        x="120"
        y="210"
        width="14"
        height="14"
        fill="var(--color-accent)"
        transform="rotate(-14 127 217)"
      />
      <rect
        x="260"
        y="90"
        width="14"
        height="14"
        fill="var(--color-surface)"
        transform="rotate(30 267 97)"
      />
      <rect
        x="320"
        y="180"
        width="16"
        height="16"
        fill="var(--color-mustard)"
        transform="rotate(-38 328 188)"
      />
      {/* podium, cropped at the bottom edge */}
      <rect x="450" y="330" width="110" height="90" fill="var(--color-turquoise)" />
      <rect x="560" y="280" width="140" height="140" fill="var(--color-surface)" />
      <rect x="700" y="340" width="120" height="80" fill="var(--color-accent)" />
      {/* the pennant on the center step */}
      <rect x="622" y="110" width="14" height="176" rx="7" fill="var(--color-surface)" />
      <circle cx="629" cy="102" r="15" fill="var(--color-mustard)" />
      <path d="M 636 126 L 810 176 L 636 228 Z" fill="var(--color-accent)" />
    </svg>
  );
}
