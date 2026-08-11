import { Link } from "react-router";
import { PublicAuthShell } from "../../components/layout/PublicAuthShell";

export function RegisterPage() {
  return (
    <PublicAuthShell
      context="Участникам"
      title="Регистрация"
      lead="Профиль участника — первый шаг к команде и подаче решения на чемпионат."
      secondary={<Link to="/login">Уже есть профиль — вход</Link>}
    />
  );
}
