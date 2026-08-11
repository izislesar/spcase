import { motion, type Variants } from "motion/react";

/*
 * The result motif: a three-step podium with the coral cup on the center
 * step and flat confetti above. Designed for navy backgrounds; steps crop
 * at the bottom edge.
 *
 * Hover hooks: the podium group can rise slightly while the two confetti
 * groups separate subtly — a single finite gesture, never a looping
 * celebration. Without variants the scene renders statically (the current
 * placement on /jury/login is static).
 */
export function PodiumScene({
  className,
  podiumVariants,
  confettiLeftVariants,
  confettiRightVariants,
}: {
  className?: string;
  podiumVariants?: Variants;
  confettiLeftVariants?: Variants;
  confettiRightVariants?: Variants;
}) {
  return (
    <svg className={className} viewBox="0 0 520 320" aria-hidden="true" focusable="false">
      {/* confetti: plain rotated rectangles, not sparks */}
      <motion.g variants={confettiLeftVariants}>
        <rect
          x="60"
          y="40"
          width="18"
          height="18"
          fill="var(--color-mustard)"
          transform="rotate(18 69 49)"
        />
        <rect
          x="120"
          y="90"
          width="14"
          height="14"
          fill="var(--color-surface)"
          transform="rotate(-24 127 97)"
        />
        <rect
          x="180"
          y="30"
          width="16"
          height="16"
          fill="var(--color-turquoise)"
          transform="rotate(40 188 38)"
        />
      </motion.g>
      <motion.g variants={confettiRightVariants}>
        <rect
          x="420"
          y="50"
          width="18"
          height="18"
          fill="var(--color-accent)"
          transform="rotate(-16 429 59)"
        />
        <rect
          x="462"
          y="110"
          width="14"
          height="14"
          fill="var(--color-mustard)"
          transform="rotate(30 469 117)"
        />
        <rect
          x="372"
          y="120"
          width="12"
          height="12"
          fill="var(--color-surface)"
          transform="rotate(-35 378 126)"
        />
      </motion.g>
      <motion.g variants={podiumVariants}>
        {/* podium */}
        <rect x="70" y="190" width="130" height="130" fill="var(--color-turquoise)" />
        <rect x="200" y="140" width="130" height="180" fill="var(--color-surface)" />
        <rect x="330" y="210" width="120" height="110" fill="var(--color-mustard)" />
        {/* the cup on the center step */}
        <path
          d="M 232 72 C 206 72 206 102 236 104 M 298 72 C 324 72 324 102 294 104"
          fill="none"
          stroke="var(--color-accent)"
          strokeWidth="10"
          strokeLinecap="round"
        />
        <path d="M 232 62 H 298 L 290 122 H 240 Z" fill="var(--color-accent)" />
        <rect x="259" y="122" width="12" height="14" fill="var(--color-ink)" />
        <rect x="238" y="136" width="54" height="12" rx="6" fill="var(--color-ink)" />
      </motion.g>
    </svg>
  );
}
