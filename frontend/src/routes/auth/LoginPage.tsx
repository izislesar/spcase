import { Link } from "react-router";
import { PublicAuthShell } from "../../components/layout/PublicAuthShell";

export function LoginPage() {
  return (
    <PublicAuthShell
      context="Участникам"
      title="Вход"
      lead="Вход для участников чемпионата — в личный кабинет команды и к отправке решения."
      secondary={<Link to="/register">Нет профиля — регистрация</Link>}
    />
  );
}
