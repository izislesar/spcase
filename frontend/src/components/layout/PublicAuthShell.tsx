import { motion, useMotionValue, useReducedMotion, useSpring, useTransform } from "motion/react";
import type { ReactNode } from "react";
import { useLocation, useViewTransitionState } from "react-router";
import { EDITORIAL_EASE, useFinePointer } from "../../lib/motion";
import styles from "./PublicAuthShell.module.css";

type Field = "mustard" | "turquoise" | "navy" | "accent";

/* CSS-module class lookups type as string | undefined; fall back to "". */
const FIELD_CLASS: Record<Field, string> = {
  mustard: styles.fieldMustard ?? "",
  turquoise: styles.fieldTurquoise ?? "",
  navy: styles.fieldNavy ?? "",
  accent: styles.fieldAccent ?? "",
};

/*
 * Visual shell for the public auth routes (/register, /login,
 * /jury/register, /jury/login): the same editorial world as the homepage —
 * display typography on the canvas, one large artwork field beside the
 * content. These routes are still functional placeholders until the auth
 * stage: the note states that plainly instead of faking a form.
 *
 * Motion: the title rises through a mask, copy follows, and the artwork
 * field enters with a short settle; fine pointers get a restrained scene
 * depth (±4px). When the route is part of a view transition, the field
 * carries its vt-name (the coral/turquoise surface morph) and the Motion
 * entrances step aside — the transition is the entrance.
 */
export function PublicAuthShell({
  eyebrow,
  title,
  lead,
  field,
  art,
  vtName,
}: {
  eyebrow: string;
  title: string;
  lead: string;
  field: Field;
  art: ReactNode;
  /** View-transition name for the field while this route transitions. */
  vtName?: "vt-coral" | "vt-turquoise";
}) {
  const reduced = useReducedMotion();
  const finePointer = useFinePointer();
  const { pathname } = useLocation();
  const transitioning = useViewTransitionState(pathname);
  const noEntrance = reduced || transitioning;
  const pointerEnabled = finePointer && !reduced;

  const pointerX = useMotionValue(0);
  const pointerY = useMotionValue(0);
  const smoothX = useSpring(pointerX, { stiffness: 60, damping: 16 });
  const smoothY = useSpring(pointerY, { stiffness: 60, damping: 16 });
  const artX = useTransform(smoothX, [-1, 1], [-4, 4]);
  const artY = useTransform(smoothY, [-1, 1], [-3, 3]);

  const onPointerMove = (event: React.PointerEvent<HTMLElement>) => {
    if (!pointerEnabled) return;
    const rect = event.currentTarget.getBoundingClientRect();
    pointerX.set(((event.clientX - rect.left) / rect.width) * 2 - 1);
    pointerY.set(((event.clientY - rect.top) / rect.height) * 2 - 1);
  };
  const onPointerLeave = () => {
    pointerX.set(0);
    pointerY.set(0);
  };

  return (
    <section
      className={styles.page}
      aria-labelledby="auth-heading"
      onPointerMove={onPointerMove}
      onPointerLeave={onPointerLeave}
    >
      <div className={`container-wide ${styles.inner}`}>
        <div className={styles.content}>
          <motion.p
            className={styles.eyebrow}
            initial={noEntrance ? false : { opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.45, ease: EDITORIAL_EASE }}
          >
            {eyebrow}
          </motion.p>
          <h1 id="auth-heading" className={styles.title}>
            <span className={styles.titleMask}>
              <motion.span
                className={styles.titleLine}
                initial={noEntrance ? false : { y: "112%" }}
                animate={{ y: "0%" }}
                transition={{ duration: 0.6, ease: EDITORIAL_EASE, delay: 0.08 }}
              >
                {title}
              </motion.span>
            </span>
          </h1>
          <motion.p
            className={styles.lead}
            initial={noEntrance ? false : { opacity: 0, y: 14 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: EDITORIAL_EASE, delay: 0.18 }}
          >
            {lead}
          </motion.p>
          <motion.p
            className={styles.note}
            initial={noEntrance ? false : { opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: EDITORIAL_EASE, delay: 0.26 }}
          >
            Рабочая форма появится на следующем этапе миграции — страница уже занимает своё место в
            структуре сайта.
          </motion.p>
        </div>
        <motion.div
          className={`${styles.field} ${FIELD_CLASS[field]}`}
          style={{ viewTransitionName: transitioning && vtName ? vtName : undefined }}
          initial={noEntrance ? false : { opacity: 0, y: 28 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.65, ease: EDITORIAL_EASE, delay: 0.15 }}
        >
          <motion.div
            className={styles.fieldArt}
            style={pointerEnabled ? { x: artX, y: artY } : undefined}
          >
            {art}
          </motion.div>
        </motion.div>
      </div>
    </section>
  );
}
