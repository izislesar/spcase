import {
  LayoutGroup,
  motion,
  useMotionValue,
  useMotionValueEvent,
  useReducedMotion,
  useScroll,
  useSpring,
  useTransform,
} from "motion/react";
import { useEffect, useId, useRef, useState } from "react";
import { NavLink, Outlet } from "react-router";
import { Mark } from "../../components/graphics/Mark";
import { MARKER_SPRING } from "../../lib/motion";
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

/*
 * Bottom bar: the active marker TRAVELS between destinations via a
 * shared-layout Motion span (navigation continuity), and the bar answers
 * meaningful scroll direction by shifting PARTIALLY downward (never fully
 * hidden). Hover, keyboard focus and the page top pin it fully visible.
 * Scroll work stays on motion values — no React state per scroll frame.
 */
function BottomBar() {
  const reduced = useReducedMotion();
  const { scrollY } = useScroll();
  const shiftTarget = useMotionValue(0);
  const shift = useSpring(shiftTarget, { stiffness: 320, damping: 34 });
  const y = useTransform(shift, [0, 1], ["0%", "62%"]);
  const pinnedRef = useRef(false);
  const directionRef = useRef<"down" | "up">("up");

  useMotionValueEvent(scrollY, "change", (latest) => {
    if (reduced || pinnedRef.current || latest < 96) {
      directionRef.current = "up";
      shiftTarget.set(0);
      return;
    }
    const delta = latest - (scrollY.getPrevious() ?? latest);
    if (delta > 3 && directionRef.current !== "down") {
      directionRef.current = "down";
      shiftTarget.set(1);
    } else if (delta < -3 && directionRef.current !== "up") {
      directionRef.current = "up";
      shiftTarget.set(0);
    }
  });

  const pin = (pinned: boolean) => {
    pinnedRef.current = pinned;
    if (pinned) {
      directionRef.current = "up";
      shiftTarget.set(0);
    }
  };

  return (
    <motion.nav
      className={styles.bottomBar}
      aria-label="Основная навигация"
      style={{ y }}
      onHoverStart={() => pin(true)}
      onHoverEnd={() => pin(false)}
      onFocusCapture={() => pin(true)}
      onBlurCapture={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget as Node | null)) pin(false);
      }}
    >
      <NavLink
        to="/"
        viewTransition
        className={styles.barBrand}
        aria-label="СПК кейс-чемпионат, на главную"
      >
        <Mark className={styles.barMark} />
        СПК
      </NavLink>
      <LayoutGroup id="bottom-nav">
        <div className={styles.barLinks}>
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              viewTransition
              className={navLinkClass}
            >
              {({ isActive }) => (
                <>
                  {isActive && (
                    <motion.span
                      layoutId="bottom-nav-marker"
                      className={styles.barMarker}
                      transition={MARKER_SPRING}
                      aria-hidden="true"
                    />
                  )}
                  {item.label}
                </>
              )}
            </NavLink>
          ))}
        </div>
      </LayoutGroup>
    </motion.nav>
  );
}

/*
 * Footer: plain meta — the factual brand line and secondary navigation on
 * the same dark canvas, separated by one hairline. No closing poster.
 */
function Footer({ menuOpen }: { menuOpen: boolean }) {
  return (
    <footer className={styles.footer} inert={menuOpen}>
      <div className={`container-wide ${styles.footerInner}`}>
        <p className={styles.footerBrand}>СПК кейс-чемпионат · Санкт-Петербург · 2026</p>
        <nav className={styles.footerNav} aria-label="Дополнительная навигация">
          <NavLink to="/register" viewTransition>
            Регистрация
          </NavLink>
          <NavLink to="/schedule" viewTransition>
            Расписание
          </NavLink>
          <NavLink to="/jury/login" viewTransition>
            Вход для жюри
          </NavLink>
        </nav>
      </div>
    </footer>
  );
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
      {/*
        Compact mobile header. On desktop it yields to the bottom bar below —
        the top of the page stays free for the content.
      */}
      <header className={styles.header}>
        <div className={`container ${styles.headerInner}`}>
          <NavLink to="/" className={styles.brand} aria-label="СПК кейс-чемпионат, на главную">
            <Mark className={styles.brandMark} />
            СПК
          </NavLink>
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
                  viewTransition
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
        {/*
          The page wrapper carries the vt-page view-transition name: route
          view transitions crossfade this region while the shared chrome
          (header/footer/bottom bar) stays put.
        */}
        <div className={styles.routePage}>
          <Outlet />
        </div>
      </main>
      <Footer menuOpen={menuOpen} />
      <BottomBar />
    </div>
  );
}
