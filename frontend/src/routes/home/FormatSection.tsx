import styles from "./FormatSection.module.css";

/*
 * The championship format as the real three stages in reading order. On
 * wide viewports the stages sit on shallow stepped material planes (Z1):
 * each later stage advances a few pixels in depth with a slightly stronger
 * separation shadow, so the sequence reads as a physical progression
 * rather than three independent cards. The stages are not interactive —
 * no hover spectacle. On smaller viewports and without the transforms the
 * list collapses to the same conventional vertical sequence with hairline
 * separators; sequence numbers stay quiet metadata, never decoration.
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
