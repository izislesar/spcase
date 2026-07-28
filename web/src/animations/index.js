import { ScrollTrigger } from "./gsap.js";
import { initSmoothScroll } from "./scroll.js";
import { initTransitions } from "./transitions.js";

const MOTION_REFRESH_EVENT = "spcase:motion:refresh";
let activeCleanup = null;

export function refreshMotion(root = document) {
  window.dispatchEvent(new CustomEvent(MOTION_REFRESH_EVENT, {
    detail: { root }
  }));
}

export function initMotionSystem() {
  if (activeCleanup) return activeCleanup;

  const stopSmoothScroll = initSmoothScroll();
  const stopTransitions = initTransitions();
  let active = true;

  const cleanup = () => {
    if (!active) return;
    active = false;
    window.removeEventListener("pagehide", handlePageHide);
    stopTransitions();
    stopSmoothScroll();
    activeCleanup = null;
  };

  const handlePageHide = (event) => {
    // Keep the live animation state when the browser stores the page in bfcache.
    if (!event.persisted) cleanup();
  };

  window.addEventListener("pagehide", handlePageHide);
  activeCleanup = cleanup;

  refreshMotion();
  requestAnimationFrame(() => ScrollTrigger.refresh());

  return cleanup;
}
