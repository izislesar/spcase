import { Block, CaseArc, Disc, HalfDisc } from "../../components/graphics/grammar";
import { ArrowLink } from "../../components/ui/ActionLinks";
import { ErrorNotice, LoadingState } from "../../components/ui/DataState";
import { errorMessage, eventDateTimeAttr, formatEventTime, useSchedule } from "./api";
import styles from "./home.module.css";

function padIndex(value: number): string {
  return String(value).padStart(2, "0");
}

/*
 * Timeline markers echo the team elements of the graphic grammar: each
 * event is marked by one form (disc / block / half-disc), cycling by
 * position. Decorative → the svg is aria-hidden.
 */
function TimelineMarker({ position }: { position: number }) {
  const kind = position % 3;
  return (
    <svg className={styles.timelineMarker} viewBox="0 0 24 24" aria-hidden="true" focusable="false">
      {kind === 0 && <Disc cx={12} cy={12} r={10} fill="var(--color-accent)" />}
      {kind === 1 && <Block cx={12} cy={12} size={19} rotate={8} fill="var(--color-on-field)" />}
      {kind === 2 && <HalfDisc cx={12} cy={14} r={10} fill="var(--color-accent)" />}
    </svg>
  );
}

export function SchedulePreview() {
  const schedule = useSchedule();

  return (
    <section className={styles.schedule} aria-labelledby="schedule-preview-heading">
      {/*
        Compositional anchor: one large case ring, still open — the schedule
        is the road toward closing it. Kept quiet (low-contrast stroke).
      */}
      <svg
        className={styles.scheduleAnchor}
        viewBox="0 0 480 480"
        aria-hidden="true"
        focusable="false"
      >
        <CaseArc
          cx={240}
          cy={240}
          r={190}
          width={34}
          gap={80}
          rotate={-60}
          stroke="color-mix(in oklch, var(--color-on-field) 13%, transparent)"
        />
      </svg>
      <div className={`container ${styles.scheduleInner}`}>
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
                  <TimelineMarker position={index} />
                  <div className={styles.timelineMeta}>
                    <span className={styles.timelineIndex} aria-hidden="true">
                      {padIndex(index + 1)}
                    </span>
                    <time
                      className={styles.timelineTime}
                      dateTime={eventDateTimeAttr(event.start_time)}
                    >
                      {formatEventTime(event.start_time)}
                    </time>
                  </div>
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
