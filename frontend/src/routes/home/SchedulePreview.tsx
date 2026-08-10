import { motion, useReducedMotion, type Variants } from "motion/react";
import { useViewTransitionState } from "react-router";
import { PublicStatus } from "../../components/public/PublicStatus";
import { ArrowLink } from "../../components/ui/ActionLinks";
import { EDITORIAL_EASE, useNarrowViewport, VIEWPORT_ONCE } from "../../lib/motion";
import { errorMessage, eventDateTimeAttr, formatEventTime, useSchedule } from "./api";
import styles from "./SchedulePreview.module.css";

/*
 * Homepage schedule preview on the turquoise field: concise typographic
 * rows — a big time stamp, the event, nothing boxed. Live data, loading,
 * error and empty states run through PublicStatus; the full program lives
 * on /schedule. The section heading carries the vt-title view-transition
 * name, so it travels into the /schedule page title when navigating.
 */

const rowVariants: Variants = {
  hidden: (narrow: boolean) => ({ opacity: 0, y: narrow ? 12 : 18 }),
  visible: { opacity: 1, y: 0, transition: { duration: 0.5, ease: EDITORIAL_EASE } },
};

export function SchedulePreview() {
  const schedule = useSchedule();
  const reduced = useReducedMotion();
  const narrow = useNarrowViewport();
  const vtSchedule = useViewTransitionState("/schedule");

  return (
    <section className={styles.schedule} aria-labelledby="schedule-preview-heading">
      <div className={`container-wide ${styles.scheduleInner}`}>
        <header className={styles.scheduleHeader}>
          <motion.p
            className={styles.eyebrow}
            initial={reduced ? false : { opacity: 0, y: 10 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={VIEWPORT_ONCE}
            transition={{ duration: 0.45, ease: EDITORIAL_EASE }}
          >
            02 · По времени
          </motion.p>
          <motion.h2
            id="schedule-preview-heading"
            className={styles.sectionTitle}
            style={{ viewTransitionName: vtSchedule ? "vt-title" : undefined }}
            initial={reduced ? false : { opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={VIEWPORT_ONCE}
            transition={{ duration: 0.6, ease: EDITORIAL_EASE, delay: 0.06 }}
          >
            Расписание
          </motion.h2>
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
            <ol className={styles.eventRows}>
              {schedule.data.events.map((event) => (
                <motion.li
                  key={event.id}
                  className={styles.eventRow}
                  initial={reduced ? false : "hidden"}
                  whileInView="visible"
                  viewport={{ once: true, amount: 0.4 }}
                  custom={narrow}
                  variants={rowVariants}
                >
                  <time className={styles.eventTime} dateTime={eventDateTimeAttr(event.start_time)}>
                    {formatEventTime(event.start_time)}
                  </time>
                  <div className={styles.eventBody}>
                    <h3 className={styles.eventTitle}>{event.title}</h3>
                    <p className={styles.eventDescription}>{event.description}</p>
                  </div>
                </motion.li>
              ))}
            </ol>
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
