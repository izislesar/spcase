import styles from "./home.module.css";

const STAGES = [
  {
    index: "01",
    name: "Старт",
    alias: "Регистрация",
    text: "Создайте профиль, соберите команду или присоединитесь по восьмизначному коду.",
  },
  {
    index: "02",
    name: "Интенсив",
    alias: "Работа",
    text: "Разберите задачу партнёра, сформулируйте подход и подготовьте одно итоговое решение.",
  },
  {
    index: "03",
    name: "Финал",
    alias: "Защита",
    text: "Представьте результат экспертному жюри и получите оценку по шести критериям.",
  },
] as const;

export function FormatSection() {
  return (
    <section className={styles.format} aria-labelledby="format-heading">
      <div className="container">
        <header className={styles.sectionHeader}>
          <p className={styles.eyebrow}>01 · Формат</p>
          <h2 id="format-heading" className={styles.sectionTitle}>
            Три этапа. Одна сильная работа.
          </h2>
          <p className={styles.sectionIntro}>
            Путь команды от регистрации до финальной защиты построен как единый интенсив.
          </p>
        </header>
        <div className={styles.stages}>
          <ProgressionPath />
          <ol className={styles.stageList}>
            {STAGES.map((stage) => (
              <li key={stage.index} className={styles.stage}>
                <span className={styles.stageNumber} aria-hidden="true">
                  {stage.index}
                </span>
                <h3 className={styles.stageName}>
                  {stage.name} <span className={styles.stageAlias}>/ {stage.alias}</span>
                </h3>
                <p className={styles.stageText}>{stage.text}</p>
              </li>
            ))}
          </ol>
        </div>
      </div>
    </section>
  );
}

/*
 * Signature interaction: the progression path connecting the three stages.
 * The stroke draws itself via a CSS scroll-driven animation as the section
 * crosses the viewport (no JS). The fully drawn path is the baseline:
 * without `animation-timeline` support or under prefers-reduced-motion the
 * animation simply does not apply and the static path is shown.
 * Marker dots sit exactly on the quadratic path joints (T-commands pass
 * through their endpoints). Both variants are decorative → aria-hidden.
 */
function ProgressionPath() {
  return (
    <>
      <svg
        className={`${styles.path} ${styles.pathHorizontal}`}
        viewBox="0 0 1200 96"
        aria-hidden="true"
        focusable="false"
      >
        <path
          data-testid="progression-path-horizontal"
          className={styles.pathLine}
          d="M 24 66 Q 112 66 200 60 T 600 36 T 1000 60 T 1176 56"
          pathLength={100}
          fill="none"
        />
        <circle cx="200" cy="60" r="6" fill="var(--color-accent)" />
        <circle cx="600" cy="36" r="6" fill="var(--color-accent)" />
        <circle cx="1000" cy="60" r="6" fill="var(--color-accent)" />
      </svg>
      <svg
        className={`${styles.path} ${styles.pathVertical}`}
        viewBox="0 0 48 640"
        preserveAspectRatio="none"
        aria-hidden="true"
        focusable="false"
      >
        <path
          data-testid="progression-path-vertical"
          className={styles.pathLine}
          d="M 24 16 Q 22 56 20 96 T 28 320 T 20 544 T 24 620"
          pathLength={100}
          fill="none"
          vectorEffect="non-scaling-stroke"
        />
        <circle cx="20" cy="96" r="6" fill="var(--color-accent)" />
        <circle cx="28" cy="320" r="6" fill="var(--color-accent)" />
        <circle cx="20" cy="544" r="6" fill="var(--color-accent)" />
      </svg>
    </>
  );
}
