import { PublicAuthShell } from "../../components/layout/PublicAuthShell";

export function JuryLoginPage() {
  return (
    <PublicAuthShell
      context="Жюри"
      title="Вход для жюри"
      lead="Вход для членов экспертного жюри — к списку команд и оценке решений."
    />
  );
}
