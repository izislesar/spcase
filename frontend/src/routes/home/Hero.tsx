import { ArrowLink, ButtonLink } from "../../components/ui/ActionLinks";
import { ErrorNotice, LoadingState } from "../../components/ui/DataState";
import { errorMessage, formatEventTime, usePublicInfo } from "./api";
import styles from "./Hero.module.css";
import { formatCountdown, useCountdown } from "./useCountdown";

/*
 * The hero as plain championship identity: the name, one factual lead
 * sentence, the primary action and a short fact list — place/year, team
 * format, jury criteria and the live registration state (countdown plus the
 * deadline date from /info, with loading, error and terminal states
 * intact). No artwork, no entrance choreography: the static composition is
 * the design.
 */
export function Hero() {
  return (
    <section className={styles.hero} aria-labelledby="home-heading">
      <div className={`container-wide ${styles.heroInner}`}>
        <h1 id="home-heading" className={styles.title}>
          <span className={styles.titleLine}>СПК</span>
          <span className={styles.titleLine}>кейс-чемпионат</span>
        </h1>
        <p className={styles.lead}>
          Практический кейс, командная работа и защита решения перед экспертами.
        </p>
        <div className={styles.heroActions}>
          <ButtonLink to="/register" viewTransition>
            Подать заявку
          </ButtonLink>
          <ArrowLink to="/schedule" viewTransition>
            Смотреть расписание
          </ArrowLink>
          <ArrowLink to="/login" viewTransition>
            Вход для участников
          </ArrowLink>
        </div>
        <dl className={styles.facts}>
          <div className={styles.fact}>
            <dt className={styles.factLabel}>Место</dt>
            <dd className={styles.factValue}>Санкт-Петербург · 2026</dd>
          </div>
          <div className={styles.fact}>
            <dt className={styles.factLabel}>Команда</dt>
            <dd className={styles.factValue}>02—04 участника</dd>
          </div>
          <div className={styles.fact}>
            <dt className={styles.factLabel}>Оценка жюри</dt>
            <dd className={styles.factValue}>6 критериев</dd>
          </div>
          <RegistrationState />
        </dl>
      </div>
    </section>
  );
}

function RegistrationState() {
  const info = usePublicInfo();
  const countdown = useCountdown(info.data?.registration_deadline);

  return (
    <div className={styles.fact}>
      <dt className={styles.factLabel}>Регистрация</dt>
      <dd className={styles.factValue}>
        {info.isPending && <LoadingState label="Загружаем дедлайн…" />}
        {info.isError && <ErrorNotice message={errorMessage(info.error)} />}
        {info.isSuccess &&
          (countdown.finished ? (
            <strong role="status">Регистрация завершена</strong>
          ) : (
            <span role="status">
              {/* Digits tick every second and are hidden from AT; the
                  accessible text changes at most once an hour. */}
              <span aria-hidden="true">
                до {formatEventTime(info.data.registration_deadline)} ·{" "}
                {formatCountdown(countdown.remainingMs ?? 0)}
              </span>
              <span className="visually-hidden">
                {accessibleRemaining(countdown.remainingMs ?? 0)}
              </span>
            </span>
          ))}
      </dd>
    </div>
  );
}

function accessibleRemaining(ms: number): string {
  const days = Math.floor(ms / 86_400_000);
  const hours = Math.floor((ms % 86_400_000) / 3_600_000);
  return `Осталось ${days} дн. ${hours} ч.`;
}
