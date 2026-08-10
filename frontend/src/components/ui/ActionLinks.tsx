import { type MotionProps, motion } from "motion/react";
import type { ComponentProps, ComponentType } from "react";
import { Link, type LinkProps } from "react-router";
import { SNAPPY_SPRING } from "../../lib/motion";
import styles from "./ActionLinks.module.css";

/*
 * Shared action microinteractions: the primary CTA gets an immediate press
 * (spring compression, no magnetic effects); editorial arrow links move
 * their arrow independently on hover. Hover styling itself stays in CSS;
 * Motion adds the physical response.
 *
 * motion.create's own prop typing collides with DOM gesture/animation
 * handlers inside LinkProps; the cast keeps the full Link contract (to,
 * viewTransition, aria…) plus the Motion gesture/variant props, dropping
 * only the handler keys whose DOM and Motion signatures are incompatible.
 */
type MotionConflictingKeys =
  | "onAnimationStart"
  | "onAnimationEnd"
  | "onAnimationIteration"
  | "onTransitionEnd"
  | "onDragStart"
  | "onDragEnd"
  | "onDrag";

export const MotionLink = motion.create(Link) as unknown as ComponentType<
  Omit<LinkProps, MotionConflictingKeys> & MotionProps
>;

type MotionLinkProps = ComponentProps<typeof MotionLink>;

/** Primary action: a flat accent block — editorial, not a pill. */
export function ButtonLink({ className, ...props }: MotionLinkProps) {
  return (
    <MotionLink
      className={`${styles.button} ${className ?? ""}`}
      whileTap={{ scale: 0.97 }}
      transition={SNAPPY_SPRING}
      {...props}
    />
  );
}

interface ArrowLinkProps extends MotionLinkProps {
  /** Arrow glyph shown after the label; decorative, hidden from AT. */
  arrow?: string;
}

const arrowVariants = {
  hover: { x: 4, transition: SNAPPY_SPRING },
};

/** Secondary action: quiet text link with an underline and a trailing arrow. */
export function ArrowLink({ className, arrow = "→", children, ...props }: ArrowLinkProps) {
  return (
    <MotionLink
      className={`${styles.arrowLink} ${className ?? ""}`}
      whileHover="hover"
      transition={SNAPPY_SPRING}
      {...props}
    >
      {children}
      <motion.span className={styles.arrow} variants={arrowVariants} aria-hidden="true">
        {arrow}
      </motion.span>
    </MotionLink>
  );
}
