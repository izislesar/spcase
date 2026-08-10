import { PodiumScene } from "../../components/graphics/scenes/PodiumScene";
import { PublicAuthShell } from "../../components/layout/PublicAuthShell";

export function JuryLoginPage() {
  return (
    <PublicAuthShell
      eyebrow="Жюри"
      title="Вход для жюри"
      lead="Вход для членов экспертного жюри — к списку команд и оценке решений."
      field="navy"
      art={<PodiumScene />}
    />
  );
}
