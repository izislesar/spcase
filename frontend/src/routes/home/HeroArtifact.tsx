import { motion, useMotionValue, useReducedMotion, useSpring, useTransform } from "motion/react";
import { useEffect, useRef } from "react";
import { TILT_SPRING } from "../../lib/motion";
import { formatEventTime, usePublicInfo } from "./api";
import styles from "./HeroArtifact.module.css";

/** Pointer tilt amplitude in degrees — a few degrees, never a performance. */
const TILT_DEGREES = 3;
/*
 * Tilt runs only where the spatial composition exists: wide viewports with
 * a fine pointer (same breakpoint as the shell's DESKTOP_MEDIA). On small
 * viewports the object is a flat panel and never tilts.
 */
const TILT_MEDIA = "(width > 64rem) and (hover: hover) and (pointer: fine)";

function clampUnit(value: number): number {
  return Math.min(1, Math.max(-1, value));
}

/*
 * The signature spatial object (Z2): the championship's own facts carried
 * on a small stack of hard graphite plates — event identity, participation
 * facts and the registration deadline on one elevated layer. Every value is
 * a truth the page already owns (static facts or live /info data); the
 * plates only organize it spatially, so the object participates in the
 * information hierarchy instead of decorating an empty region.
 *
 * Pointer tilt (±3° through one slow spring) is bound only for fine
 * pointers when motion is allowed; coordinates live in motion values, never
 * in React state. Reduced motion and coarse pointers get the static layered
 * composition; small viewports get one flat panel (see the module CSS). The
 * live ticking countdown with role="status" stays in the hero fact list —
 * the deadline here is a plain static date line, not a second live region.
 */
export function HeroArtifact() {
  const reduced = useReducedMotion();
  const sceneRef = useRef<HTMLDivElement>(null);
  const pointerX = useMotionValue(0);
  const pointerY = useMotionValue(0);
  const rotateX = useSpring(
    useTransform(pointerY, [-1, 1], [TILT_DEGREES, -TILT_DEGREES]),
    TILT_SPRING,
  );
  const rotateY = useSpring(
    useTransform(pointerX, [-1, 1], [-TILT_DEGREES, TILT_DEGREES]),
    TILT_SPRING,
  );

  useEffect(() => {
    const scene = sceneRef.current;
    if (reduced || !scene || !window.matchMedia(TILT_MEDIA).matches) return;

    const onPointerMove = (event: PointerEvent) => {
      const rect = scene.getBoundingClientRect();
      pointerX.set(clampUnit(((event.clientX - rect.left) / rect.width) * 2 - 1));
      pointerY.set(clampUnit(((event.clientY - rect.top) / rect.height) * 2 - 1));
    };
    const onPointerLeave = () => {
      pointerX.set(0);
      pointerY.set(0);
    };

    scene.addEventListener("pointermove", onPointerMove);
    scene.addEventListener("pointerleave", onPointerLeave);
    return () => {
      scene.removeEventListener("pointermove", onPointerMove);
      scene.removeEventListener("pointerleave", onPointerLeave);
    };
  }, [reduced, pointerX, pointerY]);

  return (
    <div className={styles.scene} ref={sceneRef}>
      <div className={styles.stack}>
        <motion.div className={styles.tilt} style={{ rotateX, rotateY }}>
          {/* The dark back plate gives the stack physical thickness. */}
          <div className={styles.plateBack} aria-hidden="true" />
          <div className={styles.plateIdentity}>
            <p className={styles.identityName}>СПК · 2026</p>
            <p className={styles.identityCity}>Кейс-чемпионат, Санкт-Петербург</p>
          </div>
          <dl className={styles.plateFacts}>
            <div className={styles.factRow}>
              <dt className={styles.factLabel}>Команда</dt>
              <dd className={styles.factValue}>02—04 участника</dd>
            </div>
            <div className={styles.factRow}>
              <dt className={styles.factLabel}>Оценка жюри</dt>
              <dd className={styles.factValue}>6 критериев</dd>
            </div>
          </dl>
          <RegistrationPlate />
        </motion.div>
      </div>
    </div>
  );
}

/*
 * The elevated metadata layer: the registration deadline as a static,
 * truthful line. The server flag wins when present; the deadline date is
 * the fallback. Pending/error collapse to a neutral dash — the hero fact
 * list below carries the live status, loading and error messaging.
 */
function RegistrationPlate() {
  const info = usePublicInfo();

  let value = "—";
  if (info.isSuccess) {
    const deadline = info.data.registration_deadline;
    const open = info.data.is_registration_open ?? new Date(deadline).getTime() > Date.now();
    value = open ? `до ${formatEventTime(deadline)}` : "завершена";
  }

  return (
    <div className={styles.plateMeta}>
      <span className={styles.metaTick} aria-hidden="true" />
      <p className={styles.metaLabel}>Регистрация</p>
      <p className={styles.metaValue}>{value}</p>
    </div>
  );
}
