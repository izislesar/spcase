/*
 * Questions: two oversized conversation bubbles — the navy one cropped by
 * the right edge, the coral one answering from below. Deliberately large
 * and overlapping; a quiet composition for the calm FAQ section.
 */
export function BubbleScene({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 420 380" aria-hidden="true" focusable="false">
      <rect x="110" y="36" width="350" height="190" rx="52" fill="var(--color-navy)" />
      <path d="M 150 226 L 132 276 L 196 232 Z" fill="var(--color-navy)" />
      <circle cx="190" cy="131" r="11" fill="var(--color-surface)" />
      <circle cx="250" cy="131" r="11" fill="var(--color-surface)" />
      <circle cx="310" cy="131" r="11" fill="var(--color-surface)" />
      <rect x="24" y="216" width="230" height="124" rx="40" fill="var(--color-accent)" />
      <path d="M 220 340 L 244 380 L 254 336 Z" fill="var(--color-accent)" />
      <circle cx="94" cy="278" r="10" fill="var(--color-surface)" />
      <circle cx="138" cy="278" r="10" fill="var(--color-surface)" />
      <circle cx="182" cy="278" r="10" fill="var(--color-surface)" />
    </svg>
  );
}
