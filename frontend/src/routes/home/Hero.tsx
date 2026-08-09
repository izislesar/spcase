import { Block, CaseArc, Disc, HalfDisc } from "../../components/graphics/grammar";
import { ArrowLink, ButtonLink } from "../../components/ui/ActionLinks";
import { ErrorNotice, LoadingState } from "../../components/ui/DataState";
import { errorMessage, usePublicInfo } from "./api";
import styles from "./home.module.css";
import { formatCountdown, useCountdown } from "./useCountdown";

export function Hero() {
  return (
    <section className={styles.hero} aria-labelledby="home-heading">
      <div className={`container ${styles.heroInner}`}>
        <p className={styles.eyebrow}>Санкт-Петербург · 2026</p>
        {/*
          The h1 is the compositional center of the poster: the oversized
          first line carries an inline team disc as its full stop, the second
          line is offset and set in the accent. The case ring behind it stays
          DOM typography — no letters are converted to SVG.
        */}
        <h1 id="home-heading" className={styles.display}>
          <span className={styles.displayTop}>
            СПК
            <svg
              className={styles.displayDot}
              viewBox="0 0 100 100"
              aria-hidden="true"
              focusable="false"
            >
              <circle cx="50" cy="50" r="48" fill="var(--color-accent)" />
            </svg>
          </span>
          <span className={styles.displayMain}>кейс-чемпионат</span>
        </h1>
        {/*
          Hero graphic: the unresolved case (open ink ring, gap at the top)
          with three independent team elements scattered around it — an accent
          disc approaching the gap, an ink block and a sun half-disc. This is
          state 0 of the grammar story that the format section continues.
          Decorative → aria-hidden.
        */}
        <svg
          className={styles.heroGraphic}
          viewBox="0 0 480 440"
          aria-hidden="true"
          focusable="false"
        >
          <CaseArc
            cx={292}
            cy={204}
            r={148}
            width={26}
            gap={62}
            rotate={-30}
            stroke="var(--color-ink)"
          />
          <Disc cx={446} cy={104} r={30} fill="var(--color-accent)" />
          <Block
            cx={120}
            cy={368}
            size={50}
            rotate={12}
            fill="var(--color-ink)"
            className={styles.heroFormBlock}
          />
          <HalfDisc cx={372} cy={352} r={32} rotate={-16} fill="var(--color-sun)" />
        </svg>
        <div className={styles.heroLead}>
          <p className={styles.lead}>
            Практический кейс, командная работа и защита решения перед экспертами.
          </p>
          <div className={styles.heroActions}>
            <ButtonLink to="/register">Подать заявку</ButtonLink>
            <ArrowLink to="/schedule">Смотреть расписание</ArrowLink>
          </div>
          <p className={styles.heroFact}>
            <strong>02—04</strong> участника в команде
          </p>
        </div>
        <Countdown />
      </div>
    </section>
  );
}

function Countdown() {
  const info = usePublicInfo();
  const countdown = useCountdown(info.data?.registration_deadline);

  return (
    <div className={styles.countdown}>
      <span className={styles.countdownLabel}>До конца регистрации</span>
      {info.isPending && <LoadingState label="Загружаем дедлайн…" />}
      {info.isError && <ErrorNotice message={errorMessage(info.error)} />}
      {info.isSuccess &&
        (countdown.finished ? (
          <strong className={styles.countdownValue} role="status">
            Регистрация завершена
          </strong>
        ) : (
          <span className={styles.countdownValue} role="status">
            {/* Digits tick every second and are hidden from AT; the accessible
                text changes at most once an hour. */}
            <span aria-hidden="true">{formatCountdown(countdown.remainingMs ?? 0)}</span>
            <span className="visually-hidden">
              {accessibleRemaining(countdown.remainingMs ?? 0)}
            </span>
          </span>
        ))}
    </div>
  );
}

function accessibleRemaining(ms: number): string {
  const days = Math.floor(ms / 86_400_000);
  const hours = Math.floor((ms % 86_400_000) / 3_600_000);
  return `Осталось ${days} дн. ${hours} ч.`;
}
