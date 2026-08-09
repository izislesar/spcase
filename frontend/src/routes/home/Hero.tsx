import { ArrowLink, ButtonLink } from "../../components/ui/ActionLinks";
import { ErrorNotice, LoadingState } from "../../components/ui/DataState";
import { errorMessage, usePublicInfo } from "./api";
import styles from "./home.module.css";
import { formatCountdown, useCountdown } from "./useCountdown";

export function Hero() {
  return (
    <section className={styles.hero} aria-labelledby="home-heading">
      <div className={`container ${styles.heroInner}`}>
        <div className={styles.heroText}>
          <p className={styles.eyebrow}>Санкт-Петербург · 2026</p>
          <h1 id="home-heading" className={styles.display}>
            СПК кейс-чемпионат
          </h1>
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
        {/*
          Hero graphic convention: original flat geometric SVG, decorative
          (aria-hidden), colors come from tokens. Three ascending steps read
          as the championship progression — the accent marks the final.
        */}
        <svg
          className={styles.heroGraphic}
          viewBox="0 0 320 300"
          aria-hidden="true"
          focusable="false"
        >
          <circle cx="160" cy="140" r="118" fill="var(--color-tint)" />
          <rect x="52" y="206" width="56" height="70" fill="var(--color-ink)" />
          <rect x="132" y="158" width="56" height="118" fill="var(--color-line)" />
          <rect x="212" y="102" width="56" height="174" fill="var(--color-accent)" />
          <rect x="228" y="56" width="20" height="20" fill="var(--color-ink)" />
        </svg>
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
