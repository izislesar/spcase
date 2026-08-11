import { PublicStatus } from "../../components/public/PublicStatus";
import { ArrowLink } from "../../components/ui/ActionLinks";
import { errorMessage, eventDateTimeAttr, formatEventTime, useSchedule } from "./api";
import styles from "./SchedulePreview.module.css";

/*
 * Homepage schedule preview: the information is the visual — an aligned
 * time column, event titles, hairline rules. Live data, loading, error and
 * empty states run through PublicStatus; the full program lives on
 * /schedule. No artwork, no entrance motion.
 */
export function SchedulePreview() {
  const schedule = useSchedule();

  return (
    <section className={styles.schedule} aria-labelledby="schedule-preview-heading">
      <div className={`container-wide ${styles.scheduleInner}`}>
        <header className={styles.scheduleHeader}>
          <h2 id="schedule-preview-heading" className={styles.sectionTitle}>
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
