import { motion, useReducedMotion, useScroll, useSpring, type Variants } from "motion/react";
import { useRef } from "react";
import { useViewTransitionState } from "react-router";
import { ClosingScene } from "../../components/graphics/scenes/ClosingScene";
import { RouteScene } from "../../components/graphics/scenes/RouteScene";
import { PublicStatus } from "../../components/public/PublicStatus";
import { EDITORIAL_EASE, SNAPPY_SPRING } from "../../lib/motion";
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
 * Dedicated schedule: an information timeline, not a scrolling story.
 * Oversized dates sit off a single ink line; the coral progress line fills
 * with page scroll (a Motion useScroll scaleY on the timeline — the plain
 * full line is the reduced-motion baseline) and day markers activate as
 * their segment reaches the reader: motion encodes temporal position, not
 * spectacle. Each event settles in one quiet reveal; the header art and
 * the closing band are static compositions. Data order is never animated.
 *
 * The h1 carries the vt-title view-transition name: arriving from the
 * homepage the title morphs from the preview heading, and the Motion
 * entrance is skipped (the transition IS the entrance).
 */

const itemVariants: Variants = {
  hidden: { opacity: 0, y: 18 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.5, ease: EDITORIAL_EASE } },
};

const markerVariants: Variants = {
  hidden: { scale: 0 },
  visible: { scale: 1, transition: { ...SNAPPY_SPRING } },
};

export function SchedulePage() {
  const schedule = useSchedule();
  const reduced = useReducedMotion();
  const vtSchedule = useViewTransitionState("/schedule");
  const noEntrance = reduced || vtSchedule;

  const timelineRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress: timelineProgress } = useScroll({
    target: timelineRef,
    offset: ["start 0.72", "end 0.55"],
  });
  const lineScaleY = useSpring(timelineProgress, { stiffness: 130, damping: 26 });

  return (
    <section className={styles.page} aria-labelledby="schedule-heading">
      <div className="container-wide">
        <header className={styles.header}>
          <div className={styles.headerCopy}>
            <p className={styles.eyebrow}>СПК кейс-чемпионат · 2026</p>
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
            <p className={styles.intro}>
              Этапы и дедлайны чемпионата. Все даты приходят с сервера и отображаются в часовом
              поясе участника.
            </p>
          </div>
          {/* The route motif: static — the schedule as a path, no drift. */}
          <div className={styles.headerArtWrap}>
            <RouteScene className={styles.headerArt} />
          </div>
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
                      variants={itemVariants}
                    >
                      <motion.span
                        className={styles.marker}
                        variants={markerVariants}
                        aria-hidden="true"
                      />
                      <time className={styles.date} dateTime={eventDateTimeAttr(event.start_time)}>
                        {parts ? (
                          <>
                            <span className={styles.day}>{parts.day}</span>
                            <span className={styles.month}>{parts.month}</span>
                            <span className={styles.time}>{parts.time}</span>
                          </>
                        ) : (
                          <span className={styles.month}>—</span>
                        )}
                      </time>
                      <div className={styles.body}>
                        <h2 className={styles.eventTitle}>{event.title}</h2>
                        <p className={styles.eventDescription}>{event.description}</p>
                      </div>
                    </motion.li>
                  );
                })}
              </ol>
              {/* The closing poster: the finish scene on its navy band. */}
              <motion.div
                className={styles.closingBand}
                initial={reduced ? false : { opacity: 0, y: 24 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true, amount: 0.3 }}
                transition={{ duration: 0.6, ease: EDITORIAL_EASE }}
              >
                <ClosingScene className={styles.closingArt} />
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
