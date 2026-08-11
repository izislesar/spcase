import { motion, useReducedMotion, type Variants } from "motion/react";
import { GearScene } from "../../components/graphics/scenes/GearScene";
import { PennantScene } from "../../components/graphics/scenes/PennantScene";
import { EDITORIAL_EASE, useNarrowViewport, VIEWPORT_ONCE } from "../../lib/motion";
import styles from "./FormatSection.module.css";

/*
 * The format story as ONE ruled progression, not three cards: a hairline
 * axis (horizontal on desktop, vertical on mobile) carries three stages of
 * deliberately different compositional weight:
 *   01 — a dominant typographic block, its oversized numeral breaking the
 *        progression rule;
 *   02 — a narrower, lower-offset block with the gear scene as a small
 *        inset (the work);
 *   03 — the only color field: a coral band that escapes the container on
 *        the right, the final's pennant breaking its top edge.
 * Color expresses hierarchy (the final is the goal); the other stages are
 * plain type and rules on the canvas. The ordered list keeps the reading
 * order intact at every viewport.
 *
 * Motion is minimal: the axis draws once (a progression rule) and the whole
 * progression settles in a single quiet one-shot reveal. No per-stage
 * choreography. Reduced motion shows the static composition immediately.
 */
const STAGES = [
  { index: "01", name: "Старт", alias: "Регистрация" },
  { index: "02", name: "Интенсив", alias: "Работа" },
  { index: "03", name: "Финал", alias: "Защита" },
] as const;

const axisVariants: Variants = {
  hidden: (narrow: boolean) => (narrow ? { scaleY: 0 } : { scaleX: 0 }),
  visible: { scaleX: 1, scaleY: 1, transition: { duration: 0.7, ease: EDITORIAL_EASE } },
};

const revealVariants: Variants = {
  hidden: { opacity: 0, y: 16 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.5, ease: EDITORIAL_EASE, delay: 0.15 } },
};

export function FormatSection() {
  const reduced = useReducedMotion();
  const narrow = useNarrowViewport();

  return (
    <section className={styles.format} aria-labelledby="format-heading">
      <div className={`container-wide ${styles.formatInner}`}>
        <header className={styles.formatHeader}>
          <p className={styles.eyebrow}>01 · Формат</p>
          <h2 id="format-heading" className={styles.formatTitle}>
            Три этапа. Одна сильная работа.
          </h2>
          <p className={styles.formatIntro}>
            Путь команды от регистрации до финальной защиты построен как единый интенсив.
          </p>
        </header>
        <motion.div
          className={styles.stageWrap}
          initial={reduced ? false : "hidden"}
          whileInView="visible"
          viewport={VIEWPORT_ONCE}
          custom={narrow}
        >
          {/* The progression rule: it draws once, the numerals break it. It
              lives outside the list — only <li> is valid inside <ol>. */}
          <motion.span className={styles.axis} variants={axisVariants} aria-hidden="true" />
          <motion.ol className={styles.stageList} variants={revealVariants}>
            <li className={styles.stageStart}>
              <span className={styles.stageNumber} aria-hidden="true">
                {STAGES[0].index}
              </span>
              <h3 className={styles.stageName}>
                {STAGES[0].name} <span className={styles.stageAlias}>/ {STAGES[0].alias}</span>
              </h3>
              <p className={styles.stageText}>
                Создайте профиль, соберите команду или присоединитесь по восьмизначному коду.
              </p>
            </li>
            <li className={styles.stageWork}>
              <span className={styles.stageNumber} aria-hidden="true">
                {STAGES[1].index}
              </span>
              <h3 className={styles.stageName}>
                {STAGES[1].name} <span className={styles.stageAlias}>/ {STAGES[1].alias}</span>
              </h3>
              <p className={styles.stageText}>
                Разберите задачу партнёра, сформулируйте подход и подготовьте одно итоговое решение.
              </p>
              {/* The work, as a small inset — not a poster. */}
              <GearScene className={styles.stageWorkArt} />
            </li>
            <li className={styles.stageFinal}>
              <div className={styles.stageFinalCopy}>
                <span className={styles.stageNumber} aria-hidden="true">
                  {STAGES[2].index}
                </span>
                <h3 className={styles.stageName}>
                  {STAGES[2].name} <span className={styles.stageAlias}>/ {STAGES[2].alias}</span>
                </h3>
                <p className={styles.stageText}>
                  Представьте результат экспертному жюри и получите оценку по шести критериям.
                </p>
              </div>
              <PennantScene className={styles.stageFinalArt} />
            </li>
          </motion.ol>
        </motion.div>
      </div>
    </section>
  );
}
