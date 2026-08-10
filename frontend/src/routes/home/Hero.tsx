import {
  motion,
  useMotionValue,
  useReducedMotion,
  useScroll,
  useSpring,
  useTransform,
  type Variants,
} from "motion/react";
import { useRef } from "react";
import { HeroScene, type HeroSceneDepth } from "../../components/graphics/scenes/HeroScene";
import { ArrowLink, ButtonLink } from "../../components/ui/ActionLinks";
import { ErrorNotice, LoadingState } from "../../components/ui/DataState";
import { EDITORIAL_EASE, useFinePointer, useNarrowViewport } from "../../lib/motion";
import { errorMessage, usePublicInfo } from "./api";
import styles from "./Hero.module.css";
import { formatCountdown, useCountdown } from "./useCountdown";

/*
 * The hero as a page-scale poster: headline block, artwork and body are
 * three separate regions of one grid. On desktop the artwork occupies the
 * whole right half and bleeds off the viewport edge; on mobile it moves
 * between the headline and the body as a full-bleed strip. Typography and
 * illustration carry comparable visual weight.
 *
 * Authored entrance (~1.3s, never a uniform opacity stagger): the eyebrow
 * enters quietly, «СПК» rises through a vertical mask, «кейс-чемпионат»
 * slides laterally through its own mask, the scene masses assemble (driven
 * inside HeroScene), and the body settles last. Reduced motion shows the
 * final composition immediately.
 */

const eyebrowVariants: Variants = {
  hidden: { opacity: 0, y: 10 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.45, ease: EDITORIAL_EASE } },
};

const topLineVariants: Variants = {
  hidden: { y: "112%" },
  visible: { y: "0%", transition: { duration: 0.65, ease: EDITORIAL_EASE, delay: 0.12 } },
};

const mainLineVariants: Variants = {
  hidden: { x: "-104%" },
  visible: { x: "0%", transition: { duration: 0.7, ease: EDITORIAL_EASE, delay: 0.3 } },
};

const bodyVariants: Variants = {
  hidden: { opacity: 0, y: 16 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.55, ease: EDITORIAL_EASE, delay: 0.55 } },
};

export function Hero() {
  const reduced = useReducedMotion();
  const finePointer = useFinePointer();
  const narrow = useNarrowViewport();
  const pointerEnabled = finePointer && !reduced;
  const scrollDepthEnabled = !narrow && !reduced;

  const heroRef = useRef<HTMLElement>(null);

  /*
   * Pointer depth: normalized pointer position drives per-layer motion
   * values through a soft spring. Magnitudes stay in the single pixels —
   * rear field ±3, machine ±5, documents ±9, pennant ±6 — no rotation,
   * nothing on touch devices, no React state per frame.
   */
  const pointerX = useMotionValue(0);
  const pointerY = useMotionValue(0);
  const smoothX = useSpring(pointerX, { stiffness: 56, damping: 14 });
  const smoothY = useSpring(pointerY, { stiffness: 56, damping: 14 });

  const fieldPX = useTransform(smoothX, [-1, 1], [-3, 3]);
  const fieldPY = useTransform(smoothY, [-1, 1], [-2, 2]);
  const machinePX = useTransform(smoothX, [-1, 1], [-5, 5]);
  const machinePY = useTransform(smoothY, [-1, 1], [-3, 3]);
  const docsPX = useTransform(smoothX, [-1, 1], [-9, 9]);
  const docsPY = useTransform(smoothY, [-1, 1], [-6, 6]);
  const flagPX = useTransform(smoothX, [-1, 1], [-6, 6]);
  const flagPY = useTransform(smoothY, [-1, 1], [-4, 4]);

  /*
   * Scroll depth: typography drifts slightly up while the artwork sinks at
   * a different rate; the rear field adds relative depth inside the scene.
   * No pinning — native scroll stays authoritative.
   */
  const { scrollYProgress } = useScroll({
    target: heroRef,
    offset: ["start start", "end start"],
  });
  const headScrollY = useTransform(scrollYProgress, [0, 1], [0, -28]);
  const artScrollY = useTransform(scrollYProgress, [0, 1], [0, 20]);
  const fieldScrollY = useTransform(scrollYProgress, [0, 1], [0, 36]);

  const fieldX = useTransform(() => (pointerEnabled ? fieldPX.get() : 0));
  const fieldY = useTransform(
    () => (pointerEnabled ? fieldPY.get() : 0) + (scrollDepthEnabled ? fieldScrollY.get() : 0),
  );

  const depthStyles: HeroSceneDepth | undefined =
    pointerEnabled || scrollDepthEnabled
      ? {
          field: { x: fieldX, y: fieldY },
          machine: pointerEnabled ? { x: machinePX, y: machinePY } : undefined,
          docs: pointerEnabled ? { x: docsPX, y: docsPY } : undefined,
          flag: pointerEnabled ? { x: flagPX, y: flagPY } : undefined,
        }
      : undefined;

  const onPointerMove = (event: React.PointerEvent<HTMLElement>) => {
    if (!pointerEnabled) return;
    const rect = event.currentTarget.getBoundingClientRect();
    pointerX.set(((event.clientX - rect.left) / rect.width) * 2 - 1);
    pointerY.set(((event.clientY - rect.top) / rect.height) * 2 - 1);
  };
  const onPointerLeave = () => {
    pointerX.set(0);
    pointerY.set(0);
  };

  return (
    <motion.section
      ref={heroRef}
      className={styles.hero}
      aria-labelledby="home-heading"
      initial={reduced ? false : "hidden"}
      animate="visible"
      onPointerMove={onPointerMove}
      onPointerLeave={onPointerLeave}
    >
      <div className={`container-wide ${styles.heroInner}`}>
        <motion.div
          className={styles.heroHead}
          style={scrollDepthEnabled ? { y: headScrollY } : undefined}
        >
          <motion.p className={styles.eyebrow} variants={eyebrowVariants}>
            Санкт-Петербург · 2026
          </motion.p>
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
        </motion.div>
        {/*
          The one hero artwork: the case-solving machine as a wide editorial
          scene. Documents fly in, the machine works them over, the pennant
          comes out. The groups assemble independently (variants inside
          HeroScene). Decorative → the svg is aria-hidden.
        */}
        <motion.div
          className={styles.heroArtWrap}
          style={scrollDepthEnabled ? { y: artScrollY } : undefined}
        >
          <HeroScene className={styles.heroArt} depthStyles={depthStyles} />
        </motion.div>
        <motion.div className={styles.heroBody} variants={bodyVariants}>
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
          <p className={styles.heroFact}>
            <strong>02—04</strong> участника в команде
          </p>
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
