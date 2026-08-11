import type { Easing, SpringOptions, Transition } from "motion/react";

/*
 * Shared motion vocabulary for the sparse effects that remain: navigation
 * continuity (the nav marker), state transition (the accordion), direct
 * interaction feedback (presses, hover nudges) and the hero object's
 * restrained spatial response. Nothing here licenses decorative
 * choreography — unique interactions keep explicit custom motion.
 */

/** Quiet ease-out: confident start, short settle. */
export const QUIET_EASE: Easing = [0.22, 0.68, 0.24, 1];

/** Snappy spring: presses and small hover affordances. */
export const SNAPPY_SPRING: Transition = { type: "spring", stiffness: 420, damping: 34 };

/** Restrained spring for the bottom-nav active marker travel. */
export const MARKER_SPRING: Transition = { type: "spring", stiffness: 360, damping: 32 };

/*
 * Slow, heavy spring for the hero fact-plate pointer tilt — the only
 * spatial pointer response in the system. Low stiffness and added mass
 * keep the few-degree movement physical rather than playful; the object
 * settles gently back to neutral.
 */
export const TILT_SPRING: SpringOptions = { stiffness: 110, damping: 22, mass: 1.2 };
