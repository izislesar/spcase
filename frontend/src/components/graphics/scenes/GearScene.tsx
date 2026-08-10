import { motion, type Variants } from "motion/react";
import { Gear } from "./Gear";

/*
 * Stage 02 — the work: two meshed gears, the large one deliberately cropped
 * by the bottom edge, and a document riding the dashed path underneath.
 * A wide horizontal composition, the opposite aspect of SheetStack.
 *
 * The gears accept optional variant sets from the parent context: the
 * format section rotates them a small finite amount once (hidden→visible),
 * the mosaic artwork field counter-rotates them while hovered (rest/hover).
 * Rotations are finite and spring-settled — never an infinite spin.
 * Without variants the scene renders statically.
 */
export function GearScene({
  className,
  largeGearVariants,
  smallGearVariants,
}: {
  className?: string;
  largeGearVariants?: Variants;
  smallGearVariants?: Variants;
}) {
  return (
    <svg className={className} viewBox="0 0 460 260" aria-hidden="true" focusable="false">
      <motion.g variants={largeGearVariants} style={{ originX: 0.5, originY: 0.5 }}>
        <Gear cx={110} cy={180} r={78} fill="var(--color-ink)" hubFill="var(--color-surface)" />
      </motion.g>
      <motion.g variants={smallGearVariants} style={{ originX: 0.5, originY: 0.5 }}>
        <Gear cx={336} cy={84} r={46} fill="var(--color-accent)" hubFill="var(--color-surface)" />
      </motion.g>
      <path
        d="M 16 232 C 140 214 260 240 444 206"
        fill="none"
        stroke="var(--color-ink)"
        strokeWidth="5"
        strokeLinecap="round"
        strokeDasharray="2 14"
      />
      <g transform="rotate(7 300 200)">
        <rect x="262" y="168" width="76" height="62" rx="6" fill="var(--color-surface)" />
        <path
          d="M 276 190 H 326 M 276 206 H 312"
          stroke="var(--color-ink)"
          strokeWidth="6"
          strokeLinecap="round"
        />
      </g>
    </svg>
  );
}
