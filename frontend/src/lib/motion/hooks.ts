import { useEffect, useState } from "react";

/*
 * Environment hooks that let choreographies choose variant sets: fine
 * pointers get hover microresponses (nothing hover-dependent ever carries
 * functionality), narrow viewports get simpler variant geometry. Reduced
 * motion is decided separately via MotionConfig (`reducedMotion="user"`)
 * and `useReducedMotion` at the call site.
 */

function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(() => window.matchMedia(query).matches);

  useEffect(() => {
    const mql = window.matchMedia(query);
    const onChange = (event: MediaQueryListEvent) => setMatches(event.matches);
    mql.addEventListener("change", onChange);
    setMatches(mql.matches);
    return () => mql.removeEventListener("change", onChange);
  }, [query]);

  return matches;
}

const FINE_POINTER = "(hover: hover) and (pointer: fine)";

/** True only for devices with a precise pointer that can hover (mouse/trackpad). */
export function useFinePointer(): boolean {
  return useMediaQuery(FINE_POINTER);
}

const NARROW_VIEWPORT = "(width <= 64rem)";

/** True below the desktop breakpoint — mobile choreography picks simpler variants. */
export function useNarrowViewport(): boolean {
  return useMediaQuery(NARROW_VIEWPORT);
}
