import { GearScene } from "../../components/graphics/scenes/GearScene";
import { PennantScene } from "../../components/graphics/scenes/PennantScene";
import { SheetStack } from "../../components/graphics/scenes/SheetStack";
import styles from "./FormatSection.module.css";

/*
 * The format story on one full-bleed mustard field — three stage pieces of
 * deliberately different anatomy, not three cards:
 *   01 — a big white sheet composition, artwork running off its bottom edge;
 *   02 — a narrow tall navy poster with the gear scene cropped inside;
 *   03 — a wide coral band, the pennant rising above its top edge.
 * No radius, no shadows: flat fields, typography and artwork do the work.
 * The ordered list keeps the reading order intact at every viewport.
 */
const STAGES = [
  { index: "01", name: "Старт", alias: "Регистрация" },
  { index: "02", name: "Интенсив", alias: "Работа" },
  { index: "03", name: "Финал", alias: "Защита" },
] as const;

export function FormatSection() {
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
        <ol className={styles.stageList}>
          <li className={styles.stageSheet}>
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
          </li>
          <li className={styles.stageGear}>
            <span className={styles.stageNumber} aria-hidden="true">
              {STAGES[1].index}
            </span>
            <h3 className={styles.stageName}>
              {STAGES[1].name} <span className={styles.stageAlias}>/ {STAGES[1].alias}</span>
            </h3>
            <p className={styles.stageText}>
              Разберите задачу партнёра, сформулируйте подход и подготовьте одно итоговое решение.
            </p>
            <GearScene className={styles.stageGearArt} />
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
        </ol>
      </div>
    </section>
  );
}
