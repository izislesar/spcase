import { HeroScene } from "../../components/graphics/scenes/HeroScene";
import { ArrowLink, ButtonLink } from "../../components/ui/ActionLinks";
import { ErrorNotice, LoadingState } from "../../components/ui/DataState";
import { errorMessage, usePublicInfo } from "./api";
import styles from "./Hero.module.css";
import { formatCountdown, useCountdown } from "./useCountdown";

/*
 * The hero as a page-scale poster: headline block, artwork and body are
 * three separate regions of one grid. On desktop the artwork occupies the
 * whole right half and bleeds off the viewport edge; on mobile it moves
 * between the headline and the body as a full-bleed strip. Typography and
 * illustration carry comparable visual weight.
 */
export function Hero() {
  return (
    <section className={styles.hero} aria-labelledby="home-heading">
      <div className={`container-wide ${styles.heroInner}`}>
        <div className={styles.heroHead}>
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
        </div>
        {/*
          The one hero artwork: the case-solving machine as a wide editorial
          scene. Documents fly in, the machine works them over, the pennant
          comes out. One-shot entrance assembly (CSS, staggered groups);
          reduced motion gets the final static composition. Decorative → the
          svg is aria-hidden.
        */}
        <div className={styles.heroArtWrap}>
          <HeroScene className={styles.heroArt} />
        </div>
        <div className={styles.heroBody}>
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
