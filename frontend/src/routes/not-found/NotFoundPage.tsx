import { Link } from "react-router";

export function NotFoundPage() {
  return (
    <section>
      <h1>Страница не найдена</h1>
      <p>Такой страницы не существует. Проверьте адрес или вернитесь на главную.</p>
      <p>
        <Link to="/">На главную</Link>
      </p>
    </section>
  );
}
