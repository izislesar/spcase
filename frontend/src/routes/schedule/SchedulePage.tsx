import {
  motion,
  useReducedMotion,
  useScroll,
  useSpring,
  useTransform,
  type Variants,
} from "motion/react";
import { useRef } from "react";
import { useViewTransitionState } from "react-router";
import { ClosingScene } from "../../components/graphics/scenes/ClosingScene";
import { RouteScene } from "../../components/graphics/scenes/RouteScene";
import { PublicStatus } from "../../components/public/PublicStatus";
import { EDITORIAL_EASE, SNAPPY_SPRING, useNarrowViewport } from "../../lib/motion";
import { errorMessage, eventDateTimeAttr, useSchedule } from "../home/api";
import styles from "./schedule.module.css";

const dayFormatter = new Intl.DateTimeFormat("ru-RU", { day: "numeric" });
const monthFormatter = new Intl.DateTimeFormat("ru-RU", { month: "long", year: "numeric" });
const timeFormatter = new Intl.DateTimeFormat("ru-RU", { hour: "2-digit", minute: "2-digit" });

function dateParts(value: string): { day: string; month: string; time: string } | null {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return null;
  return {
    day: dayFormatter.format(date),
    month: monthFormatter.format(date),
    time: timeFormatter.format(date),
  };
}

/*
 * Dedicated schedule: a large vertical storytelling page. Oversized dates
 * sit off a single ink line; the coral progress line fills with page scroll
 * (a Motion useScroll scaleY on the timeline — the plain full line is the
 * reduced-motion baseline). Day markers activate as their segment reaches
 * the reader, dates reveal through a clip, event copy follows with a short
 * delay, and the header art drifts at its own scroll depth. Data order is
 * never animated; the choreography lives on the furniture around it.
 *
 * The h1 carries the vt-title view-transition name: arriving from the
 * homepage the title morphs from the preview heading, and the Motion
 * entrance is skipped (the transition IS the entrance).
 */

const markerVariants: Variants = {
  hidden: { scale: 0 },
  visible: { scale: 1, transition: { ...SNAPPY_SPRING } },
};

const dateVariants: Variants = {
  hidden: (narrow: boolean) => ({ opacity: 0, y: narrow ? 14 : 26, clipPath: "inset(0 0 100% 0)" }),
  visible: {
    opacity: 1,
    y: 0,
    clipPath: "inset(0 0 0% 0)",
    transition: { duration: 0.6, ease: EDITORIAL_EASE },
  },
};

const bodyVariants: Variants = {
  hidden: (narrow: boolean) => ({ opacity: 0, y: narrow ? 14 : 26 }),
  visible: { opacity: 1, y: 0, transition: { duration: 0.55, ease: EDITORIAL_EASE, delay: 0.12 } },
};

export function SchedulePage() {
  const schedule = useSchedule();
  const reduced = useReducedMotion();
  const narrow = useNarrowViewport();
  const vtSchedule = useViewTransitionState("/schedule");
  const noEntrance = reduced || vtSchedule;

  const timelineRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress: timelineProgress } = useScroll({
    target: timelineRef,
    offset: ["start 0.72", "end 0.55"],
  });
  const lineScaleY = useSpring(timelineProgress, { stiffness: 130, damping: 26 });

  /* The header route art drifts at a slightly different scroll depth. */
  const { scrollYProgress: pageProgress } = useScroll();
  const artDriftY = useTransform(pageProgress, [0, 0.3], [0, 26]);

  return (
    <section className={styles.page} aria-labelledby="schedule-heading">
      <div className="container-wide">
        <header className={styles.header}>
          <div className={styles.headerCopy}>
            <motion.p
              className={styles.eyebrow}
              initial={noEntrance ? false : { opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.45, ease: EDITORIAL_EASE }}
            >
              СПК кейс-чемпионат · 2026
            </motion.p>
            <h1
              id="schedule-heading"
              className={styles.title}
              style={{ viewTransitionName: vtSchedule ? "vt-title" : undefined }}
            >
              <span className={styles.titleMask}>
                <motion.span
                  className={styles.titleLine}
                  initial={noEntrance ? false : { y: "112%" }}
                  animate={{ y: "0%" }}
                  transition={{ duration: 0.65, ease: EDITORIAL_EASE, delay: 0.1 }}
                >
                  Расписание
                </motion.span>
              </span>
            </h1>
            <motion.p
              className={styles.intro}
              initial={noEntrance ? false : { opacity: 0, y: 14 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, ease: EDITORIAL_EASE, delay: 0.22 }}
            >
              Этапы и дедлайны чемпионата. Все даты приходят с сервера и отображаются в часовом
              поясе участника.
            </motion.p>
          </div>
          <motion.div
            className={styles.headerArtWrap}
            initial={noEntrance ? false : { opacity: 0, x: -24 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.6, ease: EDITORIAL_EASE, delay: 0.3 }}
            style={reduced || narrow ? undefined : { y: artDriftY }}
          >
            <RouteScene className={styles.headerArt} />
          </motion.div>
        </header>
        {schedule.isPending && <PublicStatus kind="loading" title="Загружаем расписание…" />}
        {schedule.isError && (
          <PublicStatus
            kind="error"
            title="Не удалось загрузить расписание"
            detail={errorMessage(schedule.error)}
            onRetry={() => schedule.refetch()}
          />
        )}
        {schedule.isSuccess &&
          (schedule.data.events.length > 0 ? (
            <div className={styles.timeline} ref={timelineRef}>
              <span className={styles.line} aria-hidden="true" />
              <motion.span
                className={styles.lineProgress}
                aria-hidden="true"
                style={reduced ? undefined : { scaleY: lineScaleY }}
              />
              <ol className={styles.items}>
                {schedule.data.events.map((event, index) => {
                  const parts = dateParts(event.start_time);
                  return (
                    <motion.li
                      key={event.id}
                      className={index % 2 === 1 ? styles.itemAlt : styles.item}
                      initial={reduced ? false : "hidden"}
                      whileInView="visible"
                      viewport={{ once: true, margin: "-18% 0px -18% 0px" }}
                      custom={narrow}
                    >
                      <motion.span
                        className={styles.marker}
                        variants={markerVariants}
                        aria-hidden="true"
                      />
                      <motion.time
                        className={styles.date}
                        variants={dateVariants}
                        dateTime={eventDateTimeAttr(event.start_time)}
                      >
                        {parts ? (
                          <>
                            <span className={styles.day}>{parts.day}</span>
                            <span className={styles.month}>{parts.month}</span>
                            <span className={styles.time}>{parts.time}</span>
                          </>
                        ) : (
                          <span className={styles.month}>—</span>
                        )}
                      </motion.time>
                      <motion.div className={styles.body} variants={bodyVariants}>
                        <h2 className={styles.eventTitle}>{event.title}</h2>
                        <p className={styles.eventDescription}>{event.description}</p>
                      </motion.div>
                    </motion.li>
                  );
                })}
              </ol>
              {/* The closing poster: the finish scene on its navy band. */}
              <motion.div
                className={styles.closingBand}
                initial={reduced ? false : { clipPath: "inset(0 0 100% 0)" }}
                whileInView={{ clipPath: "inset(0 0 0% 0)" }}
                viewport={{ once: true, amount: 0.3 }}
                transition={{ duration: 0.7, ease: EDITORIAL_EASE }}
              >
                <motion.div
                  initial={reduced ? false : { opacity: 0, y: 28 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true, amount: 0.3 }}
                  transition={{ duration: 0.6, ease: EDITORIAL_EASE, delay: 0.18 }}
                >
                  <ClosingScene className={styles.closingArt} />
                </motion.div>
              </motion.div>
            </div>
          ) : (
            <div className={styles.empty}>
              <RouteScene className={styles.emptyArt} />
              <p className={styles.emptyText} role="status">
                События пока не опубликованы. Загляните позже.
              </p>
            </div>
          ))}
      </div>
    </section>
  );
}
