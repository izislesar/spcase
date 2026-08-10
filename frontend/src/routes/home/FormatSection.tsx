import { motion, useReducedMotion, type Variants } from "motion/react";
import { GearScene } from "../../components/graphics/scenes/GearScene";
import { PennantScene } from "../../components/graphics/scenes/PennantScene";
import { SheetStack } from "../../components/graphics/scenes/SheetStack";
import { EDITORIAL_EASE, useNarrowViewport, VIEWPORT_ONCE } from "../../lib/motion";
import styles from "./FormatSection.module.css";

/*
 * The format story on one full-bleed mustard field — three stage pieces of
 * deliberately different anatomy, not three cards:
 *   01 — a big white sheet composition, artwork running off its bottom edge;
 *   02 — a narrow tall navy poster with the gear scene cropped inside;
 *   03 — a wide coral band, the pennant rising above its top edge.
 * No radius, no shadows: flat fields, typography and artwork do the work.
 * The ordered list keeps the reading order intact at every viewport.
 *
 * Individualized entrance choreography (whileInView, one shot): 01 rises
 * with the sheet stack and the pencil arriving separately; 02 slides
 * laterally and its gear rotates a small finite amount once; 03 wipes in
 * through a lateral clip while the pennant settles on its own timing.
 * Mobile (custom=narrow) shortens travel; reduced motion shows the static
 * composition immediately.
 */
const STAGES = [
  { index: "01", name: "Старт", alias: "Регистрация" },
  { index: "02", name: "Интенсив", alias: "Работа" },
  { index: "03", name: "Финал", alias: "Защита" },
] as const;

const sheetStageVariants: Variants = {
  hidden: (narrow: boolean) => ({ opacity: 0, y: narrow ? 20 : 36 }),
  visible: { opacity: 1, y: 0, transition: { duration: 0.6, ease: EDITORIAL_EASE } },
};

const gearStageVariants: Variants = {
  hidden: (narrow: boolean) => ({ opacity: 0, x: narrow ? 0 : 28 }),
  visible: { opacity: 1, x: 0, transition: { duration: 0.6, ease: EDITORIAL_EASE } },
};

const finalStageVariants: Variants = {
  hidden: { clipPath: "inset(0 100% 0 0)" },
  visible: { clipPath: "inset(0 0% 0 0)", transition: { duration: 0.7, ease: EDITORIAL_EASE } },
};

const gearLargeVariants: Variants = {
  hidden: { rotate: -12 },
  visible: { rotate: 0, transition: { type: "spring", stiffness: 90, damping: 16, delay: 0.25 } },
};

const gearSmallVariants: Variants = {
  hidden: { rotate: 10 },
  visible: { rotate: 0, transition: { type: "spring", stiffness: 90, damping: 16, delay: 0.32 } },
};

export function FormatSection() {
  const reduced = useReducedMotion();
  const narrow = useNarrowViewport();

  return (
    <section className={styles.format} aria-labelledby="format-heading">
      <div className={`container-wide ${styles.formatInner}`}>
        <header className={styles.formatHeader}>
          <motion.p
            className={styles.eyebrow}
            initial={reduced ? false : { opacity: 0, y: 10 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={VIEWPORT_ONCE}
            transition={{ duration: 0.45, ease: EDITORIAL_EASE }}
          >
            01 · Формат
          </motion.p>
          <motion.h2
            id="format-heading"
            className={styles.formatTitle}
            initial={reduced ? false : { opacity: 0, y: 24, clipPath: "inset(0 0 100% 0)" }}
            whileInView={{ opacity: 1, y: 0, clipPath: "inset(0 0 0% 0)" }}
            viewport={VIEWPORT_ONCE}
            transition={{ duration: 0.65, ease: EDITORIAL_EASE, delay: 0.08 }}
          >
            Три этапа. Одна сильная работа.
          </motion.h2>
          <motion.p
            className={styles.formatIntro}
            initial={reduced ? false : { opacity: 0, y: 14 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={VIEWPORT_ONCE}
            transition={{ duration: 0.5, ease: EDITORIAL_EASE, delay: 0.18 }}
          >
            Путь команды от регистрации до финальной защиты построен как единый интенсив.
          </motion.p>
        </header>
        <ol className={styles.stageList}>
          <motion.li
            className={styles.stageSheet}
            initial={reduced ? false : "hidden"}
            whileInView="visible"
            viewport={VIEWPORT_ONCE}
            custom={narrow}
            variants={sheetStageVariants}
          >
            <div className={styles.stageSheetCopy}>
              <span className={styles.stageNumber} aria-hidden="true">
                {STAGES[0].index}
              </span>
              <h3 className={styles.stageName}>
                {STAGES[0].name} <span className={styles.stageAlias}>/ {STAGES[0].alias}</span>
              </h3>
              <p className={styles.stageText}>
                Создайте профиль, соберите команду или присоединитесь по восьмизначному коду.
              </p>
            </div>
            <SheetStack className={styles.stageSheetArt} />
          </motion.li>
          <motion.li
            className={styles.stageGear}
            initial={reduced ? false : "hidden"}
            whileInView="visible"
            viewport={VIEWPORT_ONCE}
            custom={narrow}
            variants={gearStageVariants}
          >
            <span className={styles.stageNumber} aria-hidden="true">
              {STAGES[1].index}
            </span>
            <h3 className={styles.stageName}>
              {STAGES[1].name} <span className={styles.stageAlias}>/ {STAGES[1].alias}</span>
            </h3>
            <p className={styles.stageText}>
              Разберите задачу партнёра, сформулируйте подход и подготовьте одно итоговое решение.
            </p>
            <GearScene
              className={styles.stageGearArt}
              largeGearVariants={gearLargeVariants}
              smallGearVariants={gearSmallVariants}
            />
          </motion.li>
          <motion.li
            className={styles.stageFinal}
            initial={reduced ? false : "hidden"}
            whileInView="visible"
            viewport={VIEWPORT_ONCE}
            custom={narrow}
            variants={finalStageVariants}
          >
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
          </motion.li>
        </ol>
      </div>
    </section>
  );
}
