import { motion, useReducedMotion } from "motion/react";
import type { ReactNode } from "react";
import { useLocation, useViewTransitionState } from "react-router";
import { EDITORIAL_EASE } from "../../lib/motion";
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
 * Motion: the title rises through a mask and the copy settles beside it —
 * one quiet gesture. The field and the scene are completely stable: no
 * pointer response, forms and controls must feel quiet. When the route is
 * part of a view transition, the field carries its vt-name (the
 * coral/turquoise surface morph) and the Motion entrances step aside —
 * the transition is the entrance.
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
  const { pathname } = useLocation();
  const transitioning = useViewTransitionState(pathname);
  const noEntrance = reduced || transitioning;

  return (
    <section className={styles.page} aria-labelledby="auth-heading">
      <div className={`container-wide ${styles.inner}`}>
        <div className={styles.content}>
          <p className={styles.eyebrow}>{eyebrow}</p>
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
          <motion.div
            initial={noEntrance ? false : { opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: EDITORIAL_EASE, delay: 0.2 }}
          >
            <p className={styles.lead}>{lead}</p>
            <p className={styles.note}>
              Рабочая форма появится на следующем этапе миграции — страница уже занимает своё место
              в структуре сайта.
            </p>
          </motion.div>
        </div>
        <div
          className={`${styles.field} ${FIELD_CLASS[field]}`}
          style={{ viewTransitionName: transitioning && vtName ? vtName : undefined }}
        >
          <div className={styles.fieldArt}>{art}</div>
        </div>
      </div>
    </section>
  );
}
