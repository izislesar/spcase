import { StageBulb, StageFlag, StageSheet } from "../../components/graphics/illustrations";
import styles from "./home.module.css";

const STAGES = [
  {
    index: "01",
    name: "Старт",
    alias: "Регистрация",
    text: "Создайте профиль, соберите команду или присоединитесь по восьмизначному коду.",
    art: "sheet",
  },
  {
    index: "02",
    name: "Интенсив",
    alias: "Работа",
    text: "Разберите задачу партнёра, сформулируйте подход и подготовьте одно итоговое решение.",
    art: "bulb",
  },
  {
    index: "03",
    name: "Финал",
    alias: "Защита",
    text: "Представьте результат экспертному жюри и получите оценку по шести критериям.",
    art: "flag",
  },
] as const;

function StageArt({ art }: { art: (typeof STAGES)[number]["art"] }) {
  if (art === "bulb") return <StageBulb className={styles.stageArt} />;
  if (art === "flag") return <StageFlag className={styles.stageArt} />;
  return <StageSheet className={styles.stageArt} />;
}

/*
 * The format story as one oversized graphic panel: the mustard field holds
 * the statement, and three stage posters of deliberately different sizes,
 * colors and offsets sit on it like pieces of print design — not like app
 * cards. The one-line story per stage stays readable in any order.
 */
export function FormatSection() {
  return (
    <section className={styles.format} aria-labelledby="format-heading">
      <div className="container">
        <div className={styles.formatPanel}>
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
            {STAGES.map((stage) => (
              <li
                key={stage.index}
                className={[styles.stage, styles[`stage${stage.index}`]].filter(Boolean).join(" ")}
              >
                <span className={styles.stageNumber} aria-hidden="true">
                  {stage.index}
                </span>
                <StageArt art={stage.art} />
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
