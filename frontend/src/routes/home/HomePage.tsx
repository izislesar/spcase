import { FaqPreview } from "./FaqPreview";
import { FormatSection } from "./FormatSection";
import { Hero } from "./Hero";
import { SchedulePreview } from "./SchedulePreview";

export function HomePage() {
  return (
    <>
      <Hero />
      <FormatSection />
      <SchedulePreview />
      <FaqPreview />
    </>
  );
}
