import { Link } from "react-router";
import { PublicAuthShell } from "../../components/layout/PublicAuthShell";

export function JuryRegisterPage() {
  return (
    <PublicAuthShell
      context="Жюри"
      title="Регистрация жюри"
      lead="Регистрация члена жюри по пригласительному ключу организаторов."
      secondary={<Link to="/jury/login">Уже есть профиль жюри — вход</Link>}
    />
  );
}
