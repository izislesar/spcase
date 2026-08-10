import { Link } from "react-router";
import { GearScene } from "../../components/graphics/scenes/GearScene";
import { PodiumScene } from "../../components/graphics/scenes/PodiumScene";
import styles from "./MosaicSection.module.css";

/*
 * Editorial collage: four heterogeneous surfaces plus one deliberate blank
 * region — a curated wall, not a Bento grid. Each piece has its own
 * anatomy (typographic navigation field, pure artwork field, quiet
 * hairline teaser, dark celebration scene); there is no shared tile
 * component, no shared padding/radius/shadow recipe. Only pieces with a
 * real destination are links.
 */
export function MosaicSection() {
  return (
    <section className={styles.mosaic} aria-label="Разделы чемпионата">
      <div className={`container-wide ${styles.mosaicGrid}`}>
        <Link to="/register" className={styles.pieceJoin}>
          <span className={styles.joinKicker}>Участие</span>
          <span className={styles.joinTitle}>
            Подать
            <br />
            заявку
          </span>
          <span className={styles.joinText}>
            Команды из двух—четырёх участников.
            <span className={styles.joinArrow} aria-hidden="true">
              →
            </span>
          </span>
        </Link>
        {/* Pure artwork field: the gears at full bleed, cropped by the piece. */}
        <div className={styles.pieceArt} aria-hidden="true">
          <GearScene className={styles.artScene} />
        </div>
        <Link to="/schedule" className={styles.pieceDates}>
          <span className={styles.datesKicker}>Даты и дедлайны</span>
          <span className={styles.datesTitle}>Расписание</span>
          <span className={styles.datesText}>
            Все этапы чемпионата в хронологии — на отдельной странице.
            <span className={styles.datesArrow} aria-hidden="true">
              ↗
            </span>
          </span>
        </Link>
        <div className={styles.pieceFinal}>
          <PodiumScene className={styles.finalScene} />
          <p className={styles.finalCaption}>
            Финал — защита решения перед экспертным жюри и оценка по шести критериям.
          </p>
        </div>
      </div>
    </section>
  );
}
