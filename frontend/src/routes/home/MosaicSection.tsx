import { motion, useReducedMotion, type Variants } from "motion/react";
import { useViewTransitionState } from "react-router";
import { GearScene } from "../../components/graphics/scenes/GearScene";
import { PodiumScene } from "../../components/graphics/scenes/PodiumScene";
import { MotionLink } from "../../components/ui/ActionLinks";
import {
  EDITORIAL_EASE,
  SNAPPY_SPRING,
  SOFT_SPRING,
  useFinePointer,
  VIEWPORT_ONCE_EARLY,
} from "../../lib/motion";
import styles from "./MosaicSection.module.css";

/*
 * Editorial collage: four heterogeneous surfaces plus one deliberate blank
 * region — a curated wall, not a Bento grid. Each piece has its own
 * anatomy (typographic navigation field, pure artwork field, quiet
 * hairline teaser, dark celebration scene); there is no shared tile
 * component, no shared padding/radius/shadow recipe. Only pieces with a
 * real destination are links.
 *
 * A unique interaction per surface: the coral join field shifts its
 * oversized title and presses firmly (and morphs into /register's field
 * via the vt-coral view transition); the turquoise gear art counter-rotates
 * its gears while hovered (fine pointers only, finite, spring-settled);
 * the dates teaser draws its hairline rule on entry and leads to
 * /schedule; the navy final lets the podium rise and the confetti separate
 * subtly. Touch gets press feedback without any hover dependency.
 */

const joinTitleVariants: Variants = {
  hover: { x: 6, transition: SNAPPY_SPRING },
};

const joinArrowVariants: Variants = {
  hover: { x: 5, y: -3, transition: SNAPPY_SPRING },
};

const gearLargeHoverVariants: Variants = {
  hover: { rotate: 16, transition: { ...SOFT_SPRING } },
};

const gearSmallHoverVariants: Variants = {
  hover: { rotate: -22, transition: { ...SOFT_SPRING } },
};

const datesTitleVariants: Variants = {
  hover: { x: 4, transition: SNAPPY_SPRING },
};

const datesArrowVariants: Variants = {
  hover: { x: 4, y: -4, transition: SNAPPY_SPRING },
};

const podiumHoverVariants: Variants = {
  hover: { y: -6, transition: { ...SOFT_SPRING } },
};

const confettiLeftHoverVariants: Variants = {
  hover: { x: -3, y: -4, transition: { ...SOFT_SPRING } },
};

const confettiRightHoverVariants: Variants = {
  hover: { x: 3, y: -5, transition: { ...SOFT_SPRING } },
};

const pieceEntryVariants = (delay: number): Variants => ({
  hidden: { opacity: 0, y: 20 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.55, ease: EDITORIAL_EASE, delay } },
});

export function MosaicSection() {
  const reduced = useReducedMotion();
  const finePointer = useFinePointer();
  const vtRegister = useViewTransitionState("/register");
  const vtLogin = useViewTransitionState("/login");

  return (
    <section className={styles.mosaic} aria-label="Разделы чемпионата">
      <div className={`container-wide ${styles.mosaicGrid}`}>
        <MotionLink
          to="/register"
          viewTransition
          className={styles.pieceJoin}
          style={{ viewTransitionName: vtRegister ? "vt-coral" : undefined }}
          initial={reduced ? false : "hidden"}
          whileInView="visible"
          viewport={VIEWPORT_ONCE_EARLY}
          variants={pieceEntryVariants(0)}
          whileHover={finePointer ? "hover" : undefined}
          whileTap={{ scale: 0.97 }}
        >
          <span className={styles.joinKicker}>Участие</span>
          <motion.span className={styles.joinTitle} variants={joinTitleVariants}>
            Подать
            <br />
            заявку
          </motion.span>
          <span className={styles.joinText}>
            Команды из двух—четырёх участников.
            <motion.span
              className={styles.joinArrow}
              variants={joinArrowVariants}
              aria-hidden="true"
            >
              →
            </motion.span>
          </span>
        </MotionLink>
        {/* Pure artwork field: the gears at full bleed, cropped by the piece. */}
        <motion.div
          className={styles.pieceArt}
          aria-hidden="true"
          style={{ viewTransitionName: vtLogin ? "vt-turquoise" : undefined }}
          initial={reduced ? false : "hidden"}
          whileInView="visible"
          viewport={VIEWPORT_ONCE_EARLY}
          variants={pieceEntryVariants(0.06)}
          whileHover={finePointer ? "hover" : undefined}
          whileTap={{ scale: 0.98 }}
        >
          <GearScene
            className={styles.artScene}
            largeGearVariants={gearLargeHoverVariants}
            smallGearVariants={gearSmallHoverVariants}
          />
        </motion.div>
        <MotionLink
          to="/schedule"
          viewTransition
          className={styles.pieceDates}
          initial={reduced ? false : "hidden"}
          whileInView="visible"
          viewport={VIEWPORT_ONCE_EARLY}
          variants={pieceEntryVariants(0.12)}
          whileHover={finePointer ? "hover" : undefined}
          whileTap={{ scale: 0.98 }}
        >
          <motion.span
            className={styles.datesRule}
            aria-hidden="true"
            variants={{
              hidden: { scaleX: 0 },
              visible: {
                scaleX: 1,
                transition: { duration: 0.6, ease: EDITORIAL_EASE, delay: 0.2 },
              },
            }}
            style={{ originX: 0 }}
          />
          <span className={styles.datesKicker}>Даты и дедлайны</span>
          <motion.span className={styles.datesTitle} variants={datesTitleVariants}>
            Расписание
          </motion.span>
          <span className={styles.datesText}>
            Все этапы чемпионата в хронологии — на отдельной странице.
            <motion.span
              className={styles.datesArrow}
              variants={datesArrowVariants}
              aria-hidden="true"
            >
              ↗
            </motion.span>
          </span>
        </MotionLink>
        <motion.div
          className={styles.pieceFinal}
          initial={reduced ? false : "hidden"}
          whileInView="visible"
          viewport={VIEWPORT_ONCE_EARLY}
          variants={pieceEntryVariants(0.18)}
          whileHover={finePointer ? "hover" : undefined}
          whileTap={{ scale: 0.98 }}
        >
          <PodiumScene
            className={styles.finalScene}
            podiumVariants={podiumHoverVariants}
            confettiLeftVariants={confettiLeftHoverVariants}
            confettiRightVariants={confettiRightHoverVariants}
          />
          <p className={styles.finalCaption}>
            Финал — защита решения перед экспертным жюри и оценка по шести критериям.
          </p>
        </motion.div>
      </div>
    </section>
  );
}
