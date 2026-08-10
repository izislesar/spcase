import { PennantScene } from "../../components/graphics/scenes/PennantScene";
import { PublicAuthShell } from "../../components/layout/PublicAuthShell";

export function JuryRegisterPage() {
  return (
    <PublicAuthShell
      eyebrow="Жюри"
      title="Регистрация жюри"
      lead="Регистрация члена жюри по пригласительному ключу организаторов."
      field="accent"
      art={<PennantScene />}
    />
  );
}
