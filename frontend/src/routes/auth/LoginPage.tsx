import { GearScene } from "../../components/graphics/scenes/GearScene";
import { PublicAuthShell } from "../../components/layout/PublicAuthShell";

export function LoginPage() {
  return (
    <PublicAuthShell
      eyebrow="Участникам"
      title="Вход"
      lead="Вход для участников чемпионата — в личный кабинет команды и к отправке решения."
      field="turquoise"
      art={<GearScene />}
      vtName="vt-turquoise"
    />
  );
}
