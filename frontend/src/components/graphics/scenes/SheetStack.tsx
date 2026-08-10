import { motion, type Variants } from "motion/react";
import { SNAPPY_SPRING, SOFT_SPRING } from "../../../lib/motion";

/*
 * Stage 01 — registration: a tall cropped stack of application sheets with
 * a pencil. The front sheet runs off the bottom edge and the pencil off the
 * top; no backdrop circle, the paper itself is the composition.
 *
 * Entrance choreography hooks (used by the format section's hidden→visible
 * variants): the sheet stack moves independently and the pencil arrives
 * separately from above. The pencil's static rotation stays on an outer
 * plain <g> so the animated inner group never clobbers it. Without a
 * variant ancestor the scene renders statically.
 */

const sheetsVariants: Variants = {
  hidden: (narrow: boolean) => ({ opacity: 0, y: narrow ? 24 : 44 }),
  visible: { opacity: 1, y: 0, transition: { ...SOFT_SPRING, delay: 0.12 } },
};

const pencilVariants: Variants = {
  hidden: (narrow: boolean) => ({ opacity: 0, y: narrow ? -16 : -30 }),
  visible: { opacity: 1, y: 0, transition: { ...SNAPPY_SPRING, delay: 0.32 } },
};

export function SheetStack({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 300 420" aria-hidden="true" focusable="false">
      <motion.g variants={sheetsVariants}>
        <rect
          x="-40"
          y="150"
          width="150"
          height="150"
          fill="var(--color-mustard)"
          transform="rotate(-12 35 225)"
        />
        <g transform="rotate(9 165 140)">
          <rect x="90" y="36" width="150" height="210" rx="10" fill="var(--color-turquoise)" />
        </g>
        <g transform="rotate(-7 140 250)">
          <rect x="55" y="120" width="175" height="320" rx="10" fill="var(--color-surface)" />
          <path
            d="M 80 170 H 205 M 80 202 H 205 M 80 234 H 176"
            stroke="var(--color-ink)"
            strokeWidth="9"
            strokeLinecap="round"
          />
          <path
            d="M 92 300 l 26 26 l 52 -58"
            fill="none"
            stroke="var(--color-accent)"
            strokeWidth="12"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </g>
      </motion.g>
      <g transform="rotate(38 235 60)">
        <motion.g variants={pencilVariants}>
          <rect x="224" y="-30" width="24" height="130" rx="8" fill="var(--color-accent)" />
          <path d="M 224 100 L 236 128 L 248 100 Z" fill="var(--color-ink)" />
        </motion.g>
      </g>
    </svg>
  );
}
