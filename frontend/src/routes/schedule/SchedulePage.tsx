import { PublicStatus } from "../../components/public/PublicStatus";
import { errorMessage, eventDateTimeAttr, type ScheduleEvent, useSchedule } from "../home/api";
import styles from "./schedule.module.css";

const dayFormatter = new Intl.DateTimeFormat("ru-RU", {
  day: "numeric",
  month: "long",
  year: "numeric",
});
const timeFormatter = new Intl.DateTimeFormat("ru-RU", { hour: "2-digit", minute: "2-digit" });

function formatDay(value: string): string | null {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : dayFormatter.format(date);
}

function formatTime(value: string): string | null {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : timeFormatter.format(date);
}

/** Local calendar-day key; unparseable dates group under "unknown". */
function dayKey(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "unknown";
  return `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
}

/*
 * Group events by calendar day, keeping server order inside each group and
 * the first-seen order of the days. Days with unparseable dates go last
 * under a neutral label.
 */
function groupByDay(events: ScheduleEvent[]): { label: string; events: ScheduleEvent[] }[] {
  const groups = new Map<string, { first: ScheduleEvent; events: ScheduleEvent[] }>();
  for (const event of events) {
    const key = dayKey(event.start_time);
    const group = groups.get(key);
    if (group) {
      group.events.push(event);
    } else {
      groups.set(key, { first: event, events: [event] });
    }
  }
  return [...groups.entries()]
    .sort(([a], [b]) => (a === "unknown" ? 1 : b === "unknown" ? -1 : 0))
    .map(([key, group]) => ({
      label: key === "unknown" ? "Дата уточняется" : (formatDay(group.first.start_time) ?? key),
      events: group.events,
    }));
}

/*
 * Dedicated schedule: the schedule itself is the composition — date groups,
 * an aligned time column, event titles and descriptions, hairline
 * separators. Live data with loading, error and empty states through
 * PublicStatus. No artwork, no scroll-driven decoration: nothing on this
 * page moves except data.
 */
export function SchedulePage() {
  const schedule = useSchedule();

  return (
    <section className={styles.page} aria-labelledby="schedule-heading">
      <div className="container-wide">
        <header className={styles.header}>
          <p className={styles.context}>СПК кейс-чемпионат · 2026</p>
          <h1 id="schedule-heading" className={styles.title}>
            Расписание
          </h1>
          <p className={styles.intro}>
            Этапы и дедлайны чемпионата. Все даты приходят с сервера и отображаются в часовом поясе
            участника.
          </p>
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
            <div className={styles.groups}>
              {groupByDay(schedule.data.events).map((group) => (
                <section
                  key={group.label}
                  className={styles.group}
                  aria-label={`События: ${group.label}`}
                >
                  <h2 className={styles.groupTitle}>{group.label}</h2>
                  <ol className={styles.events}>
                    {group.events.map((event) => (
                      <li key={event.id} className={styles.event}>
                        <time
                          className={styles.time}
                          dateTime={eventDateTimeAttr(event.start_time)}
                        >
                          {formatTime(event.start_time) ?? "—"}
                        </time>
                        <div className={styles.body}>
                          <h3 className={styles.eventTitle}>{event.title}</h3>
                          <p className={styles.eventDescription}>{event.description}</p>
                        </div>
                      </li>
                    ))}
                  </ol>
                </section>
              ))}
            </div>
          ) : (
            <PublicStatus
              kind="empty"
              title="События пока не опубликованы"
              detail="Организаторы опубликуют программу — даты появятся здесь."
            />
          ))}
      </div>
    </section>
  );
}
