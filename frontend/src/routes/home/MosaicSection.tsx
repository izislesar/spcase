import { Link } from "react-router";
import { CalendarArt, ChatArt, CupArt, StageSheet } from "../../components/graphics/illustrations";
import styles from "./home.module.css";

/*
 * Editorial mosaic: a small set of large heterogeneous tiles — visual
 * navigation and poster surfaces for the real public destinations and
 * content, not SaaS feature cards. Tiles differ in size, dominant color
 * and artwork; one is pure illustration. Only tiles with a real
 * destination are links.
 */
export function MosaicSection() {
  return (
    <section className={styles.mosaic} aria-label="Разделы чемпионата">
      <div className={`container ${styles.mosaicGrid}`}>
        <Link to="/register" className={`${styles.tile} ${styles.tileJoin}`}>
          <span className={styles.tileKicker}>Участие</span>
          <span className={styles.tileTitle}>Подать заявку</span>
          <span className={styles.tileText}>Команды из двух—четырёх участников.</span>
          <StageSheet className={styles.tileArt} />
        </Link>
        <Link to="/schedule" className={`${styles.tile} ${styles.tileSchedule}`}>
          <span className={styles.tileKicker}>Даты</span>
          <span className={styles.tileTitle}>Расписание</span>
          <span className={styles.tileText}>Все этапы и дедлайны чемпионата.</span>
          <CalendarArt className={styles.tileArt} />
        </Link>
        <div className={`${styles.tile} ${styles.tileFinal}`}>
          <span className={styles.tileKicker}>Защита</span>
          <span className={styles.tileTitle}>Финал и жюри</span>
          <span className={styles.tileText}>Оценка решения по шести критериям.</span>
          <CupArt className={styles.tileArt} />
        </div>
        <div className={`${styles.tile} ${styles.tileFaq}`}>
          <span className={styles.tileKicker}>Детали</span>
          <span className={styles.tileTitle}>Вопросы и ответы</span>
          <span className={styles.tileText}>Коротко о главном — ниже на странице.</span>
          <ChatArt className={styles.tileArt} />
        </div>
        {/* Pure illustration tile: the cup at large scale, no text at all. */}
        <div className={`${styles.tile} ${styles.tileArtOnly}`} aria-hidden="true">
          <CupArt className={styles.tileArtLarge} />
        </div>
      </div>
    </section>
  );
}
