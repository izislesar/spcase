import { ArrowLink } from "../../components/ui/ActionLinks";
import { ErrorNotice, LoadingState } from "../../components/ui/DataState";
import { errorMessage, eventDateTimeAttr, formatEventTime, useSchedule } from "./api";
import styles from "./home.module.css";

function padIndex(value: number): string {
  return String(value).padStart(2, "0");
}

export function SchedulePreview() {
  const schedule = useSchedule();

  return (
    <section className={styles.schedule} aria-labelledby="schedule-preview-heading">
      <div className="container">
        <header className={styles.sectionHeader}>
          <p className={`${styles.eyebrow} ${styles.eyebrowOnField}`}>02 · По времени</p>
          <h2 id="schedule-preview-heading" className={`${styles.sectionTitle} ${styles.onField}`}>
            Расписание
          </h2>
        </header>
        {schedule.isPending && <LoadingState label="Загружаем расписание…" />}
        {schedule.isError && <ErrorNotice message={errorMessage(schedule.error)} />}
        {schedule.isSuccess &&
          (schedule.data.events.length > 0 ? (
            <ol className={styles.timeline}>
              {schedule.data.events.map((event, index) => (
                <li key={event.id} className={styles.timelineItem}>
                  <span className={styles.timelineIndex} aria-hidden="true">
                    {padIndex(index + 1)}
                  </span>
                  <time
                    className={styles.timelineTime}
                    dateTime={eventDateTimeAttr(event.start_time)}
                  >
                    {formatEventTime(event.start_time)}
                  </time>
                  <div className={styles.timelineBody}>
                    <h3 className={styles.timelineTitle}>{event.title}</h3>
                    <p className={styles.timelineDescription}>{event.description}</p>
                  </div>
                </li>
              ))}
            </ol>
          ) : (
            <p className={styles.onFieldMuted}>События пока не опубликованы.</p>
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
