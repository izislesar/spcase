import Lenis from "lenis";

import { gsap, ScrollTrigger } from "./gsap.js";

const REDUCED_MOTION_QUERY = "(prefers-reduced-motion: reduce)";
const SCROLL_LOCK_EVENT = "spcase:scroll-lock";
const NATIVE_SCROLL_AREAS = [
  "[data-lenis-prevent]",
  ".modal-panel",
  ".mobile-menu__panel"
].join(", ");

export function initSmoothScroll() {
  const reducedMotion = window.matchMedia(REDUCED_MOTION_QUERY);
  const locks = new Set();
  let instance = null;
  let tickerCallback = null;

  const syncLockState = () => {
    if (!instance) return;
    if (locks.size > 0) {
      instance.stop();
      return;
    }
    instance.start();
  };

  const destroy = () => {
    if (!instance) return;

    gsap.ticker.remove(tickerCallback);
    instance.destroy();
    instance = null;
    tickerCallback = null;

    // Lenis owns the ticker setting while it is active. Restore GSAP defaults.
    gsap.ticker.lagSmoothing(500, 33);
  };

  const start = () => {
    if (instance || reducedMotion.matches) return;

    const lenis = new Lenis({
      anchors: true,
      autoRaf: false,
      autoResize: true,
      lerp: 0.1,
      smoothWheel: true,
      syncTouch: false,
      prevent: (node) => (
        node instanceof Element &&
        Boolean(node.closest(NATIVE_SCROLL_AREAS))
      )
    });
    const tick = (time) => {
      lenis.raf(time * 1000);
    };

    lenis.on("scroll", ScrollTrigger.update);
    gsap.ticker.add(tick);
    gsap.ticker.lagSmoothing(0);

    instance = lenis;
    tickerCallback = tick;
    syncLockState();

    requestAnimationFrame(() => ScrollTrigger.refresh());
  };

  const syncPreference = () => {
    if (reducedMotion.matches) {
      destroy();
      return;
    }
    start();
  };

  const updateScrollLock = (event) => {
    const source = String(event.detail?.source || "anonymous");
    if (event.detail?.locked) {
      locks.add(source);
    } else {
      locks.delete(source);
    }
    syncLockState();
  };

  reducedMotion.addEventListener("change", syncPreference);
  window.addEventListener(SCROLL_LOCK_EVENT, updateScrollLock);
  syncPreference();

  return () => {
    reducedMotion.removeEventListener("change", syncPreference);
    window.removeEventListener(SCROLL_LOCK_EVENT, updateScrollLock);
    locks.clear();
    destroy();
  };
}
