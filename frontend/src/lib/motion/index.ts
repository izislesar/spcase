import type { Easing, Transition } from "motion/react";

/*
 * Shared editorial motion vocabulary for the public frontend. This is NOT
 * an animation framework: unique scenes keep explicit custom motion. These
 * are only the common easings, springs, timings and viewport thresholds
 * that keep separate choreographies feeling like one language.
 */

/** Editorial ease-out: confident start, long quiet settle. */
export const EDITORIAL_EASE: Easing = [0.22, 0.68, 0.24, 1];

/** Snappy spring: presses, markers, small affordances. */
export const SNAPPY_SPRING: Transition = { type: "spring", stiffness: 420, damping: 34 };

/** Soft spring: large surfaces and scene masses. */
export const SOFT_SPRING: Transition = { type: "spring", stiffness: 110, damping: 18 };

/** Restrained spring for the bottom-nav active marker travel. */
export const MARKER_SPRING: Transition = { type: "spring", stiffness: 360, damping: 32 };

/** whileInView threshold: fire once when a third of the piece is visible. */
export const VIEWPORT_ONCE = { once: true, amount: 0.35 } as const;

/** Earlier one-shot trigger for small elements. */
export const VIEWPORT_ONCE_EARLY = { once: true, amount: 0.2 } as const;

export { useFinePointer, useNarrowViewport } from "./hooks";
