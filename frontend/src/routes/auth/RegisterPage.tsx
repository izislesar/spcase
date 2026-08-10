import { SheetStack } from "../../components/graphics/scenes/SheetStack";
import { PublicAuthShell } from "../../components/layout/PublicAuthShell";

export function RegisterPage() {
  return (
    <PublicAuthShell
      eyebrow="Участникам"
      title="Регистрация"
      lead="Профиль участника — первый шаг к команде и подаче решения на чемпионат."
      field="mustard"
      art={<SheetStack />}
      vtName="vt-coral"
    />
  );
}
