import { NavLink, Outlet } from "react-router";

export function AppShell() {
  return (
    <div className="shell">
      <a className="skip-link" href="#main-content">
        К содержанию
      </a>
      <header className="shell-header">
        <div className="container shell-header-inner">
          <NavLink className="shell-brand" to="/">
            СПК
          </NavLink>
          <nav className="shell-nav" aria-label="Основная навигация">
            <NavLink to="/">Главная</NavLink>
            <NavLink to="/schedule">Расписание</NavLink>
            <NavLink to="/dashboard">Кабинет</NavLink>
            <NavLink to="/login">Вход</NavLink>
          </nav>
        </div>
      </header>
      <main className="shell-main" id="main-content">
        <div className="container">
          <Outlet />
        </div>
      </main>
      <footer className="shell-footer">
        <div className="container">
          <nav className="shell-footer-nav" aria-label="Дополнительная навигация">
            <NavLink to="/register">Регистрация</NavLink>
            <NavLink to="/schedule">Расписание</NavLink>
            <NavLink to="/jury/login">Вход для жюри</NavLink>
          </nav>
        </div>
      </footer>
    </div>
  );
}
