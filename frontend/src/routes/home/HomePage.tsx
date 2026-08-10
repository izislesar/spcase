import { FaqPreview } from "./FaqPreview";
import { FormatSection } from "./FormatSection";
import { Hero } from "./Hero";
import { MosaicSection } from "./MosaicSection";
import { SchedulePreview } from "./SchedulePreview";

export function HomePage() {
  return (
    <>
      <Hero />
      <FormatSection />
      <MosaicSection />
      <SchedulePreview />
      <FaqPreview />
    </>
  );
}
