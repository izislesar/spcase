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
import { NavLink, Outlet, useLocation, useViewTransitionState } from "react-router";
import { Mark } from "../../components/graphics/Mark";
import { ClosingScene } from "../../components/graphics/scenes/ClosingScene";
import { ButtonLink } from "../../components/ui/ActionLinks";
import { EDITORIAL_EASE, MARKER_SPRING } from "../../lib/motion";
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
 * During a view transition that involves the homepage, <html> carries
 * .vt-involves-home on both snapshots; the non-home side additionally
 * carries .vt-back. The view-transitions.css rules use them to reverse the
 * page slide when the destination is home (restrained reverse).
 */
function useViewTransitionDirection() {
  const involvesHome = useViewTransitionState("/");
  const { pathname } = useLocation();

  useEffect(() => {
    const root = document.documentElement;
    root.classList.toggle("vt-involves-home", involvesHome);
    root.classList.toggle("vt-back", involvesHome && pathname !== "/");
    return () => root.classList.remove("vt-involves-home", "vt-back");
  }, [involvesHome, pathname]);
}

/*
 * Bottom bar signature: the active marker TRAVELS between destinations via
 * a shared-layout Motion span, and the bar answers meaningful scroll
 * direction by shifting PARTIALLY downward (never fully hidden). Hover,
 * keyboard focus and the page top pin it fully visible. Scroll work stays
 * on motion values — no React state per scroll frame.
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
 * Footer: the closing poster. On approach the statement reveals line by
 * line through masks, the finish scene enters laterally (a different axis),
 * and the meta row settles last. The composition itself is unchanged.
 */
function Footer({ menuOpen }: { menuOpen: boolean }) {
  const reduced = useReducedMotion();

  return (
    <footer className={styles.footer} inert={menuOpen}>
      <div className={`container-wide ${styles.footerInner}`}>
        <div className={styles.footerCta}>
          <p className={styles.footerStatement}>
            <span className={styles.statementMask}>
              <motion.span
                className={styles.statementLine}
                initial={reduced ? false : { y: "112%" }}
                whileInView={{ y: "0%" }}
                viewport={{ once: true, amount: 0.6 }}
                transition={{ duration: 0.6, ease: EDITORIAL_EASE }}
              >
                Собери команду.
              </motion.span>
            </span>
            <span className={styles.statementMask}>
              <motion.span
                className={styles.statementLine}
                initial={reduced ? false : { y: "112%" }}
                whileInView={{ y: "0%" }}
                viewport={{ once: true, amount: 0.6 }}
                transition={{ duration: 0.6, ease: EDITORIAL_EASE, delay: 0.09 }}
              >
                Реши кейс.
              </motion.span>
            </span>
          </p>
          <ButtonLink to="/register" viewTransition className={styles.footerButton}>
            Подать заявку
          </ButtonLink>
        </div>
        <motion.div
          className={styles.footerSceneWrap}
          initial={reduced ? false : { opacity: 0, x: 44 }}
          whileInView={{ opacity: 1, x: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.7, ease: EDITORIAL_EASE, delay: 0.12 }}
        >
          <ClosingScene className={styles.footerScene} />
        </motion.div>
        <motion.div
          className={styles.footerMeta}
          initial={reduced ? false : { opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true, amount: 0.6 }}
          transition={{ duration: 0.5, ease: EDITORIAL_EASE, delay: 0.2 }}
        >
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
        </motion.div>
      </div>
    </footer>
  );
}

export function AppShell() {
  const menuId = useId();
  const [menuOpen, setMenuOpen] = useState(false);
  const toggleRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);

  useViewTransitionDirection();

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
        the reference-led pattern keeps the top of the page free for the
        editorial composition.
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
          <ClosingScene className={styles.menuEcho} />
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
          view transitions slide this region while the shared chrome
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
