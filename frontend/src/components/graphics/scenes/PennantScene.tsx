/*
 * Stage 03 — the final: a tall pennant raised on a winner's step. Designed
 * for coral/accent fields: pole and steps in ink, the pennant in deep navy,
 * the knob and the cropped side field in mustard. Crops at left and bottom.
 */
export function PennantScene({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 280 440" aria-hidden="true" focusable="false">
      <circle cx="-30" cy="110" r="140" fill="var(--color-mustard)" />
      <rect x="36" y="356" width="208" height="84" fill="var(--color-ink)" />
      <rect x="96" y="306" width="88" height="50" fill="var(--color-surface)" />
      <rect x="130" y="84" width="14" height="226" rx="7" fill="var(--color-ink)" />
      <circle cx="137" cy="76" r="14" fill="var(--color-mustard)" />
      <path d="M 144 96 L 262 132 L 144 170 Z" fill="var(--color-navy)" />
    </svg>
  );
}
