import { Block, CaseArc, Disc, HalfDisc, ResolvedMark } from "../../components/graphics/grammar";
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
          <Trajectory />
          <ol className={styles.stageList}>
            {STAGES.map((stage) => (
              <li key={stage.index} className={styles.stage}>
                <span className={styles.stageNumber} aria-hidden="true">
                  {stage.index}
                </span>
                <StageVignette index={stage.index} />
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
 * The grammar story told in three states of the SAME forms: the team
 * elements (disc, block, half-disc) scatter around the open case ring,
 * converge while the ring's gap narrows, and finally assemble into the
 * resolved mark inside the now-closed ring. All decorative → aria-hidden.
 */
function StageVignette({ index }: { index: (typeof STAGES)[number]["index"] }) {
  if (index === "03") {
    return <ResolvedMark className={`${styles.stageVignette} ${styles.stageMark}`} />;
  }
  if (index === "02") {
    return (
      <svg
        className={styles.stageVignette}
        viewBox="0 0 200 160"
        aria-hidden="true"
        focusable="false"
      >
        <CaseArc cx={100} cy={80} r={44} width={9} gap={30} rotate={-8} stroke="var(--color-ink)" />
        <Disc cx={58} cy={100} r={13} fill="var(--color-accent)" />
        <Block cx={144} cy={58} size={22} rotate={45} fill="var(--color-ink)" />
        <HalfDisc cx={126} cy={116} r={14} rotate={-28} fill="var(--color-paper)" />
      </svg>
    );
  }
  return (
    <svg
      className={styles.stageVignette}
      viewBox="0 0 200 160"
      aria-hidden="true"
      focusable="false"
    >
      <CaseArc cx={100} cy={80} r={44} width={9} gap={70} rotate={-24} stroke="var(--color-ink)" />
      <Disc cx={32} cy={128} r={13} fill="var(--color-accent)" />
      <Block cx={168} cy={32} size={22} rotate={12} fill="var(--color-ink)" />
      <HalfDisc cx={162} cy={128} r={14} rotate={8} fill="var(--color-paper)" />
    </svg>
  );
}

/*
 * Signature interaction: the trajectory one team element travels between the
 * three states. Desktop draws a horizontal path with three hops; the accent
 * disc rides it via `offset-path` driven by a CSS scroll timeline (no JS,
 * no render loop). Baselines stay fully understandable without motion:
 * the path is always drawn and the disc rests at the finish — browsers
 * without offset-path, without animation-timeline or with reduced motion
 * all keep that static end state. Mobile gets a native vertical wavy path.
 * Both variants are decorative → aria-hidden.
 */
function Trajectory() {
  return (
    <>
      <svg
        className={`${styles.path} ${styles.pathHorizontal}`}
        viewBox="0 0 1200 140"
        aria-hidden="true"
        focusable="false"
      >
        <path
          data-testid="progression-path-horizontal"
          className={styles.pathLine}
          d="M 40 118 C 120 118 160 42 240 42 C 320 42 330 118 440 118 C 520 118 560 42 640 42 C 720 42 730 118 840 118 C 920 118 960 42 1040 42 C 1120 42 1140 96 1170 108"
          fill="none"
        />
        <circle className={styles.traveller} cx={1170} cy={108} r={9} fill="var(--color-accent)" />
      </svg>
      <svg
        className={`${styles.path} ${styles.pathVertical}`}
        viewBox="0 0 60 720"
        preserveAspectRatio="none"
        aria-hidden="true"
        focusable="false"
      >
        <path
          data-testid="progression-path-vertical"
          className={styles.pathLine}
          d="M 30 16 C 44 100 14 160 32 250 C 50 340 12 400 30 490 C 48 580 18 640 30 704"
          fill="none"
          vectorEffect="non-scaling-stroke"
        />
      </svg>
    </>
  );
}
