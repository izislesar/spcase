import { useEffect, useId, useRef, useState } from "react";
import { NavLink, Outlet } from "react-router";
import { BrandMark, CaseArc, Disc, ResolvedMark } from "../../components/graphics/grammar";
import styles from "./AppShell.module.css";

const NAV_ITEMS = [
  { to: "/", label: "Главная", end: true },
  { to: "/schedule", label: "Расписание", end: false },
  { to: "/dashboard", label: "Кабинет", end: false },
  { to: "/login", label: "Вход", end: false },
] as const;

const DESKTOP_MEDIA = "(width > 64rem)";

function navLinkClass({ isActive }: { isActive: boolean }): string {
  return [styles.navLink, isActive ? styles.navLinkActive : undefined].filter(Boolean).join(" ");
}

export function AppShell() {
  const menuId = useId();
  const [menuOpen, setMenuOpen] = useState(false);
  const toggleRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);

  /*
   * Mobile menu lifecycle: while open it locks page scroll, moves focus into
   * the panel, traps Tab inside it, closes on Escape (focus returns to the
   * toggle) and closes when the viewport grows past the mobile breakpoint.
   * All listeners and the scroll lock are torn down on close/unmount.
   */
  useEffect(() => {
    if (!menuOpen) return;

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    const desktopMedia = window.matchMedia(DESKTOP_MEDIA);
    const closeOnDesktop = () => {
      if (desktopMedia.matches) setMenuOpen(false);
    };
    desktopMedia.addEventListener("change", closeOnDesktop);

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        setMenuOpen(false);
        toggleRef.current?.focus();
        return;
      }
      if (event.key === "Tab" && panelRef.current) {
        const focusable = panelRef.current.querySelectorAll<HTMLElement>("a[href], button");
        const first = focusable.item(0);
        const last = focusable.item(focusable.length - 1);
        if (event.shiftKey && document.activeElement === first) {
          event.preventDefault();
          last.focus();
        } else if (!event.shiftKey && document.activeElement === last) {
          event.preventDefault();
          first.focus();
        }
      }
    };
    document.addEventListener("keydown", onKeyDown);

    panelRef.current?.querySelector<HTMLElement>("a[href], button")?.focus();

    return () => {
      document.body.style.overflow = previousOverflow;
      desktopMedia.removeEventListener("change", closeOnDesktop);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [menuOpen]);

  const closeMenu = () => setMenuOpen(false);

  return (
    <div className={styles.shell}>
      <a className="skip-link" href="#main-content">
        К содержимому
      </a>
      <header className={styles.header}>
        <div className={`container ${styles.headerInner}`}>
          <NavLink to="/" className={styles.brand} aria-label="СПК кейс-чемпионат, на главную">
            <BrandMark className={styles.brandMark} />
            СПК
          </NavLink>
          <nav className={styles.desktopNav} aria-label="Основная навигация">
            {NAV_ITEMS.map((item) => (
              <NavLink key={item.to} to={item.to} end={item.end} className={navLinkClass}>
                {item.label}
              </NavLink>
            ))}
          </nav>
          <button
            ref={toggleRef}
            type="button"
            className={styles.menuToggle}
            aria-expanded={menuOpen}
            aria-controls={menuId}
            onClick={() => setMenuOpen((open) => !open)}
          >
            <svg
              aria-hidden="true"
              width="20"
              height="14"
              viewBox="0 0 20 14"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <path d="M0 1h20M0 7h20M0 13h20" />
            </svg>
            Меню
          </button>
        </div>
      </header>
      {menuOpen && (
        <div className={styles.menuOverlay}>
          {/*
            The dark menu echoes the grammar: the still-open case ring and one
            team disc. Decorative → aria-hidden.
          */}
          <svg
            className={styles.menuEcho}
            viewBox="0 0 400 400"
            aria-hidden="true"
            focusable="false"
          >
            <CaseArc
              cx={110}
              cy={310}
              r={140}
              width={24}
              gap={70}
              rotate={-140}
              stroke="color-mix(in oklch, var(--color-on-field) 16%, transparent)"
            />
            <Disc cx={336} cy={84} r={24} fill="var(--color-accent)" />
          </svg>
          <div
            ref={panelRef}
            id={menuId}
            className={styles.menuPanel}
            role="dialog"
            aria-modal="true"
            aria-label="Меню"
          >
            <button
              type="button"
              className={styles.menuClose}
              onClick={() => {
                closeMenu();
                toggleRef.current?.focus();
              }}
            >
              Закрыть
              <svg
                aria-hidden="true"
                width="14"
                height="14"
                viewBox="0 0 14 14"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M1 1l12 12M13 1L1 13" />
              </svg>
            </button>
            <nav className={styles.mobileNav} aria-label="Мобильная навигация">
              {NAV_ITEMS.map((item) => (
                <NavLink
                  key={item.to}
                  to={item.to}
                  end={item.end}
                  className={navLinkClass}
                  onClick={closeMenu}
                >
                  {item.label}
                </NavLink>
              ))}
            </nav>
          </div>
        </div>
      )}
      <main id="main-content" tabIndex={-1} className={styles.main} inert={menuOpen}>
        <Outlet />
      </main>
      <footer className={styles.footer} inert={menuOpen}>
        <div className={`container ${styles.footerInner}`}>
          {/*
            The story completes here: the resolved mark (closed ring, assembled
            elements) tinted for the accent field.
          */}
          <ResolvedMark className={styles.footerMark} />
          <p className={styles.footerBrand}>СПК кейс-чемпионат · Санкт-Петербург · 2026</p>
          <nav className={styles.footerNav} aria-label="Дополнительная навигация">
            <NavLink to="/register">Регистрация</NavLink>
            <NavLink to="/schedule">Расписание</NavLink>
            <NavLink to="/jury/login">Вход для жюри</NavLink>
          </nav>
        </div>
      </footer>
    </div>
  );
}
