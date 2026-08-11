import type { Easing, Transition } from "motion/react";

/*
 * Shared motion vocabulary for the sparse effects that remain: navigation
 * continuity (the nav marker), state transition (the accordion) and direct
 * interaction feedback (presses, hover nudges). Nothing here licenses
 * decorative choreography — unique interactions keep explicit custom motion.
 */

/** Quiet ease-out: confident start, short settle. */
export const QUIET_EASE: Easing = [0.22, 0.68, 0.24, 1];

/** Snappy spring: presses and small hover affordances. */
export const SNAPPY_SPRING: Transition = { type: "spring", stiffness: 420, damping: 34 };

/** Restrained spring for the bottom-nav active marker travel. */
export const MARKER_SPRING: Transition = { type: "spring", stiffness: 360, damping: 32 };
