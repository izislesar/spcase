import { CalendarArt, FlagScene } from "../../components/graphics/illustrations";
import { ErrorNotice, LoadingState } from "../../components/ui/DataState";
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
 * Dedicated schedule: a large vertical storytelling timeline. Oversized
 * dates sit off a single ink line; a coral progress line fills with page
 * scroll (native CSS scroll timeline, fully guarded — the plain line is
 * the baseline). The composition intentionally occupies space, so the
 * page stays designed even with few events.
 */
export function SchedulePage() {
  const schedule = useSchedule();

  return (
    <section className={styles.page} aria-labelledby="schedule-heading">
      <div className="container">
        <header className={styles.header}>
          <div className={styles.headerCopy}>
            <p className={styles.eyebrow}>СПК кейс-чемпионат · 2026</p>
            <h1 id="schedule-heading" className={styles.title}>
              Расписание
            </h1>
            <p className={styles.intro}>
              Этапы и дедлайны чемпионата. Все даты приходят с сервера и отображаются в часовом
              поясе участника.
            </p>
          </div>
          <CalendarArt className={styles.headerArt} />
        </header>
        {schedule.isPending && <LoadingState label="Загружаем расписание…" />}
        {schedule.isError && <ErrorNotice message={errorMessage(schedule.error)} />}
        {schedule.isSuccess &&
          (schedule.data.events.length > 0 ? (
            <div className={styles.timeline}>
              <span className={styles.line} aria-hidden="true" />
              <span className={styles.lineProgress} aria-hidden="true" />
              <ol className={styles.items}>
                {schedule.data.events.map((event, index) => {
                  const parts = dateParts(event.start_time);
                  return (
                    <li key={event.id} className={index % 2 === 1 ? styles.itemAlt : styles.item}>
                      <span className={styles.marker} aria-hidden="true" />
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
                    </li>
                  );
                })}
              </ol>
              <FlagScene className={styles.closingArt} />
            </div>
          ) : (
            <div className={styles.empty}>
              <CalendarArt className={styles.emptyArt} />
              <p className={styles.emptyText}>События пока не опубликованы. Загляните позже.</p>
            </div>
          ))}
      </div>
    </section>
  );
}
