import { motion, useReducedMotion, type Variants } from "motion/react";
import { HeroScene } from "../../components/graphics/scenes/HeroScene";
import { ArrowLink, ButtonLink } from "../../components/ui/ActionLinks";
import { ErrorNotice, LoadingState } from "../../components/ui/DataState";
import { EDITORIAL_EASE } from "../../lib/motion";
import { errorMessage, usePublicInfo } from "./api";
import styles from "./Hero.module.css";
import { formatCountdown, useCountdown } from "./useCountdown";

/*
 * The hero as the cover/status board of the championship, not a startup
 * hero: an oversized masthead headline, a ruled dossier strip with the
 * factual state of the competition (place/year, team format, jury criteria,
 * live registration countdown), and the case-solving machine crossing the
 * strip's top rule — the single authored irregularity of the composition.
 * Everything on the strip is truthful: static facts or live /info data with
 * its loading, error and terminal states intact.
 *
 * Authored entrance (~1.2s, never a uniform opacity stagger): «СПК» rises
 * through a vertical mask, «кейс-чемпионат» slides laterally through its
 * own mask, the scene masses assemble (driven inside HeroScene), and the
 * body with the data strip settles last. There is no pointer or scroll
 * depth: the static composition carries the design. Reduced motion shows
 * the final composition immediately.
 */

const topLineVariants: Variants = {
  hidden: { y: "112%" },
  visible: { y: "0%", transition: { duration: 0.65, ease: EDITORIAL_EASE, delay: 0.12 } },
};

const mainLineVariants: Variants = {
  hidden: { x: "-104%" },
  visible: { x: "0%", transition: { duration: 0.7, ease: EDITORIAL_EASE, delay: 0.3 } },
};

const settleVariants: Variants = {
  hidden: { opacity: 0, y: 16 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.55, ease: EDITORIAL_EASE, delay: 0.55 } },
};

export function Hero() {
  const reduced = useReducedMotion();

  return (
    <motion.section
      className={styles.hero}
      aria-labelledby="home-heading"
      initial={reduced ? false : "hidden"}
      animate="visible"
    >
      <div className={`container-wide ${styles.heroInner}`}>
        <div className={styles.heroHead}>
          {/*
            Heavy grotesk headline with two deliberate line breaks; the second
            line is offset and set in the coral accent. Each line reveals
            through its own overflow mask — the exact text and the h1
            semantics are unchanged.
          */}
          <h1 id="home-heading" className={styles.display}>
            <span className={styles.displayTop}>
              <motion.span className={styles.displayLine} variants={topLineVariants}>
                СПК
              </motion.span>
            </span>
            <span className={styles.displayMain}>
              <motion.span className={styles.displayLine} variants={mainLineVariants}>
                кейс-чемпионат
              </motion.span>
            </span>
          </h1>
        </div>
        <motion.div className={styles.heroBody} variants={settleVariants}>
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
          </div>
        </motion.div>
        {/*
          The one hero artwork: documents fly in, the machine works them
          over, the pennant comes out. On desktop its bottom mass crosses
          the data strip's top rule into the strip's deliberately empty
          right region. Decorative → the svg is aria-hidden.
        */}
        <div className={styles.heroArtWrap}>
          <HeroScene className={styles.heroArt} />
        </div>
        {/* The dossier strip: hairline rules, small caps labels, tabular values. */}
        <motion.div className={styles.strip} variants={settleVariants}>
          <div className={styles.stripCell}>
            <span className={styles.cellLabel}>Место</span>
            <span className={styles.cellValue}>Санкт-Петербург · 2026</span>
          </div>
          <div className={styles.stripCell}>
            <span className={styles.cellLabel}>Команда</span>
            <span className={styles.cellValue}>02—04 участника</span>
          </div>
          <div className={styles.stripCell}>
            <span className={styles.cellLabel}>Оценка жюри</span>
            <span className={styles.cellValue}>6 критериев</span>
          </div>
          <Countdown />
        </motion.div>
      </div>
    </motion.section>
  );
}

function Countdown() {
  const info = usePublicInfo();
  const countdown = useCountdown(info.data?.registration_deadline);

  return (
    <div className={styles.stripCell}>
      <span className={styles.cellLabel}>До конца регистрации</span>
      {info.isPending && <LoadingState label="Загружаем дедлайн…" />}
      {info.isError && <ErrorNotice message={errorMessage(info.error)} />}
      {info.isSuccess &&
        (countdown.finished ? (
          <strong className={styles.cellData} role="status">
            Регистрация завершена
          </strong>
        ) : (
          <span className={styles.cellData} role="status">
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
