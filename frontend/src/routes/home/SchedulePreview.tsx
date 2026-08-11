import { motion, useReducedMotion, type Variants } from "motion/react";
import { useViewTransitionState } from "react-router";
import { PublicStatus } from "../../components/public/PublicStatus";
import { ArrowLink } from "../../components/ui/ActionLinks";
import { EDITORIAL_EASE, VIEWPORT_ONCE } from "../../lib/motion";
import { errorMessage, eventDateTimeAttr, formatEventTime, useSchedule } from "./api";
import styles from "./SchedulePreview.module.css";

/*
 * Homepage schedule preview on the turquoise field: scoreboard grammar —
 * a big tabular time column, hairline rules, nothing boxed. Live data,
 * loading, error and empty states run through PublicStatus; the full
 * program lives on /schedule. The section heading carries the vt-title
 * view-transition name, so it travels into the /schedule page title when
 * navigating.
 *
 * Motion: the rows settle in a single quiet one-shot reveal; the header is
 * static. Reduced motion shows everything immediately.
 */

const rowsVariants: Variants = {
  hidden: { opacity: 0, y: 14 },
  visible: { opacity: 1, y: 0, transition: { duration: 0.5, ease: EDITORIAL_EASE } },
};

export function SchedulePreview() {
  const schedule = useSchedule();
  const reduced = useReducedMotion();
  const vtSchedule = useViewTransitionState("/schedule");

  return (
    <section className={styles.schedule} aria-labelledby="schedule-preview-heading">
      <div className={`container-wide ${styles.scheduleInner}`}>
        <header className={styles.scheduleHeader}>
          <p className={styles.eyebrow}>02 · По времени</p>
          <h2
            id="schedule-preview-heading"
            className={styles.sectionTitle}
            style={{ viewTransitionName: vtSchedule ? "vt-title" : undefined }}
          >
            Расписание
          </h2>
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
            <motion.ol
              className={styles.eventRows}
              initial={reduced ? false : "hidden"}
              whileInView="visible"
              viewport={VIEWPORT_ONCE}
              variants={rowsVariants}
            >
              {schedule.data.events.map((event) => (
                <li key={event.id} className={styles.eventRow}>
                  <time className={styles.eventTime} dateTime={eventDateTimeAttr(event.start_time)}>
                    {formatEventTime(event.start_time)}
                  </time>
                  <div className={styles.eventBody}>
                    <h3 className={styles.eventTitle}>{event.title}</h3>
                    <p className={styles.eventDescription}>{event.description}</p>
                  </div>
                </li>
              ))}
            </motion.ol>
          ) : (
            <PublicStatus
              kind="empty"
              title="События пока не опубликованы"
              detail="Организаторы опубликуют программу — даты появятся здесь и на странице расписания."
            />
          ))}
        <footer className={styles.scheduleFooter}>
          <p className={styles.scheduleNote}>
            Все даты приходят с сервера и отображаются в часовом поясе участника.
          </p>
          <ArrowLink to="/schedule" viewTransition arrow="↗" className={styles.scheduleLink}>
            Полная программа
          </ArrowLink>
        </footer>
      </div>
    </section>
  );
}
