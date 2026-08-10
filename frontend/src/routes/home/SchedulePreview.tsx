import { ArrowLink } from "../../components/ui/ActionLinks";
import { ErrorNotice, LoadingState } from "../../components/ui/DataState";
import { errorMessage, eventDateTimeAttr, formatEventTime, useSchedule } from "./api";
import styles from "./SchedulePreview.module.css";

/*
 * Homepage schedule preview on the turquoise field: concise typographic
 * rows — a big time stamp, the event, nothing boxed. Live data, loading,
 * error and empty states are preserved; full program lives on /schedule.
 */
export function SchedulePreview() {
  const schedule = useSchedule();

  return (
    <section className={styles.schedule} aria-labelledby="schedule-preview-heading">
      <div className={`container-wide ${styles.scheduleInner}`}>
        <header className={styles.scheduleHeader}>
          <p className={styles.eyebrow}>02 · По времени</p>
          <h2 id="schedule-preview-heading" className={styles.sectionTitle}>
            Расписание
          </h2>
        </header>
        {schedule.isPending && <LoadingState label="Загружаем расписание…" />}
        {schedule.isError && <ErrorNotice message={errorMessage(schedule.error)} />}
        {schedule.isSuccess &&
          (schedule.data.events.length > 0 ? (
            <ol className={styles.eventRows}>
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
            </ol>
          ) : (
            <p className={styles.scheduleEmpty}>События пока не опубликованы.</p>
          ))}
        <footer className={styles.scheduleFooter}>
          <p className={styles.scheduleNote}>
            Все даты приходят с сервера и отображаются в часовом поясе участника.
          </p>
          <ArrowLink to="/schedule" arrow="↗" className={styles.scheduleLink}>
            Полная программа
          </ArrowLink>
        </footer>
      </div>
    </section>
  );
}
