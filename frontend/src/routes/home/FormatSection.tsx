import styles from "./FormatSection.module.css";

/*
 * The championship format as the real three stages in reading order. On
 * wide viewports the stages form ONE connected stepped slab (Z1): full-width
 * terraces that overlap the previous step, advance slightly in depth and
 * share one base plane, so the sequence reads as a single physical
 * progression rather than three independent panels. The stages are not
 * interactive — no hover spectacle. On smaller viewports and without the
 * transforms the list collapses to the same conventional vertical sequence
 * with hairline separators; sequence numbers stay quiet metadata, never
 * decoration.
 */
const STAGES = [
  {
    number: 1,
    name: "Старт",
    text: "Создайте профиль, соберите команду или присоединитесь по восьмизначному коду.",
  },
  {
    number: 2,
    name: "Интенсив",
    text: "Разберите задачу партнёра, сформулируйте подход и подготовьте одно итоговое решение.",
  },
  {
    number: 3,
    name: "Финал",
    text: "Представьте результат экспертному жюри и получите оценку по шести критериям.",
  },
] as const;

export function FormatSection() {
  return (
    <section className={styles.format} aria-labelledby="format-heading">
      <div className={`container-wide ${styles.formatInner}`}>
        <header className={styles.formatHeader}>
          <h2 id="format-heading" className={styles.formatTitle}>
            Формат чемпионата
          </h2>
          <p className={styles.formatIntro}>
            Путь команды от регистрации до финальной защиты построен как единый интенсив.
          </p>
        </header>
        <ol className={styles.stageList}>
          {STAGES.map((stage) => (
            <li className={styles.stage} key={stage.number}>
              <span className={styles.stageNumber}>Этап {stage.number}</span>
              <h3 className={styles.stageName}>{stage.name}</h3>
              <p className={styles.stageText}>{stage.text}</p>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}
