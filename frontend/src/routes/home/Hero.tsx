import { CaseMachine } from "../../components/graphics/illustrations";
import { ArrowLink, ButtonLink } from "../../components/ui/ActionLinks";
import { ErrorNotice, LoadingState } from "../../components/ui/DataState";
import { errorMessage, usePublicInfo } from "./api";
import styles from "./home.module.css";
import { formatCountdown, useCountdown } from "./useCountdown";

export function Hero() {
  return (
    <section className={styles.hero} aria-labelledby="home-heading">
      <div className={`container ${styles.heroInner}`}>
        <div className={styles.heroCopy}>
          <p className={styles.eyebrow}>Санкт-Петербург · 2026</p>
          {/*
            Heavy grotesk headline with two deliberate line breaks; the second
            line is offset and set in the coral accent. DOM typography only —
            the illustration interacts with it spatially, never replaces it.
          */}
          <h1 id="home-heading" className={styles.display}>
            <span className={styles.displayTop}>СПК</span>
            <span className={styles.displayMain}>кейс-чемпионат</span>
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
          <Countdown />
        </div>
        {/*
          The one hero artwork: the case-solving machine. Documents fly in,
          the machine works them over, the championship pennant comes out.
          The scene plays a one-shot entrance assembly (CSS, staggered
          groups); reduced motion and unsupported engines get the final
          static composition. Decorative → the svg is aria-hidden.
        */}
        <CaseMachine className={styles.heroArt} />
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
