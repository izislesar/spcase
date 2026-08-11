import { motion, type Variants } from "motion/react";
import { useViewTransitionState } from "react-router";
import { GearScene } from "../../components/graphics/scenes/GearScene";
import { MotionLink } from "../../components/ui/ActionLinks";
import { SNAPPY_SPRING, SOFT_SPRING, useFinePointer } from "../../lib/motion";
import { formatEventTime, usePublicInfo } from "./api";
import styles from "./MosaicSection.module.css";

/*
 * Two editorial fields, no tile wall: the registration CTA is a strong
 * coral band — oversized «Подать заявку» type above a ruled meta row of
 * truthful facts (team format; the live registration deadline from /info,
 * falling back to an honest «уточняется» until the data arrives). The
 * turquoise gear field stays as the quiet asymmetric companion: smaller,
 * lower, cropped — the login continuity element.
 *
 * Interactions are microresponses only: the band presses and its title and
 * arrow nudge on hover; the large gear turns a single small amount while
 * hovered (fine pointers only, finite, spring-settled — the signature
 * microresponse of the work motif). No entrance choreography: the view
 * transitions (vt-coral into /register, vt-turquoise into /login) provide
 * the motion that matters here.
 */

const bandTitleVariants: Variants = {
  hover: { x: 6, transition: SNAPPY_SPRING },
};

const bandArrowVariants: Variants = {
  hover: { x: 5, y: -3, transition: SNAPPY_SPRING },
};

const gearNudgeVariants: Variants = {
  hover: { rotate: 10, transition: { ...SOFT_SPRING } },
};

export function MosaicSection() {
  const info = usePublicInfo();
  const finePointer = useFinePointer();
  const vtRegister = useViewTransitionState("/register");
  const vtLogin = useViewTransitionState("/login");

  return (
    <section className={styles.mosaic} aria-label="Разделы чемпионата">
      <div className={`container-wide ${styles.mosaicInner}`}>
        <MotionLink
          to="/register"
          viewTransition
          className={styles.band}
          style={{ viewTransitionName: vtRegister ? "vt-coral" : undefined }}
          whileHover={finePointer ? "hover" : undefined}
          whileTap={{ scale: 0.98 }}
        >
          <span className={styles.bandKicker}>Участие</span>
          <motion.span className={styles.bandTitle} variants={bandTitleVariants}>
            Подать
            <br />
            заявку
          </motion.span>
          <span className={styles.bandMeta}>
            <span className={styles.metaCell}>
              <span className={styles.metaLabel}>Команда</span>
              <span className={styles.metaValue}>02—04 участника</span>
            </span>
            <span className={styles.metaCell}>
              <span className={styles.metaLabel}>Дедлайн регистрации</span>
              <span className={styles.metaValue}>
                {info.isSuccess ? formatEventTime(info.data.registration_deadline) : "уточняется"}
              </span>
            </span>
            <motion.span
              className={styles.bandArrow}
              variants={bandArrowVariants}
              aria-hidden="true"
            >
              →
            </motion.span>
          </span>
        </MotionLink>
        {/*
          Login continuity: a quiet turquoise field beside the band. It is
          not a link; when the reader navigates to /login the field morphs
          into the login page's artwork field (vt-turquoise).
        */}
        <motion.div
          className={styles.loginField}
          aria-hidden="true"
          style={{ viewTransitionName: vtLogin ? "vt-turquoise" : undefined }}
          whileHover={finePointer ? "hover" : undefined}
        >
          <GearScene className={styles.loginArt} largeGearVariants={gearNudgeVariants} />
        </motion.div>
      </div>
    </section>
  );
}
