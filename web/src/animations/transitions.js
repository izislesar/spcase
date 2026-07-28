import { gsap, ScrollTrigger } from "./gsap.js";

const MOTION_QUERY = "(prefers-reduced-motion: no-preference)";
const FINE_POINTER_QUERY = "(hover: hover) and (pointer: fine)";
const MOTION_REFRESH_EVENT = "spcase:motion:refresh";

function findWithin(root, selector) {
  const matches = [];

  if (root instanceof Element && root.matches(selector)) {
    matches.push(root);
  }
  if (root && typeof root.querySelectorAll === "function") {
    matches.push(...root.querySelectorAll(selector));
  }

  return matches;
}

function isRenderable(element) {
  return element.isConnected && element.getClientRects().length > 0;
}

function numberAttribute(element, name, fallback) {
  const value = Number.parseFloat(element.dataset[name]);
  return Number.isFinite(value) ? value : fallback;
}

function createHeroIntro(hero) {
  const label = hero.querySelector("[data-hero-label]");
  const lines = hero.querySelectorAll("[data-hero-line]");
  const copy = hero.querySelector("[data-hero-copy]");
  const actions = hero.querySelector("[data-hero-actions]");
  const layers = hero.querySelectorAll("[data-hero-layer]");
  const orbit = hero.querySelector(".hero-orbit");
  const timeline = gsap.timeline({
    defaults: {
      ease: "power3.out"
    }
  });

  if (label) {
    timeline.fromTo(
      label,
      { opacity: 0, y: 14 },
      { duration: 0.45, opacity: 1, y: 0 },
      0.05
    );
  }
  if (lines.length) {
    timeline.fromTo(
      lines,
      { yPercent: 112 },
      {
        duration: 0.98,
        ease: "power4.out",
        stagger: 0.13,
        yPercent: 0
      },
      0.34
    );
  }
  if (copy) {
    timeline.fromTo(
      copy,
      { opacity: 0, y: 24 },
      { duration: 0.75, opacity: 1, y: 0 },
      0.76
    );
  }
  if (actions) {
    timeline.fromTo(
      actions,
      { opacity: 0, y: 20 },
      { duration: 0.72, opacity: 1, y: 0 },
      0.94
    );
  }
  if (orbit) {
    timeline.fromTo(
      orbit,
      { opacity: 0, scale: 0.985 },
      { duration: 1.4, opacity: 1, scale: 1 },
      0.18
    );
  }
  if (layers.length) {
    timeline.fromTo(
      layers,
      {
        opacity: 0,
        scale: 0.9,
        y: (index) => 18 + index * 4
      },
      {
        duration: 1.05,
        ease: "power3.out",
        opacity: 1,
        scale: 1,
        stagger: 0.06,
        y: 0
      },
      0.22
    );
  }
}

function createReveal(element) {
  const restingY = Number(gsap.getProperty(element, "y")) || 0;

  gsap.fromTo(
    element,
    {
      opacity: 0,
      y: restingY + 34
    },
    {
      duration: 0.9,
      ease: "power3.out",
      opacity: 1,
      scrollTrigger: {
        once: true,
        start: "top 88%",
        trigger: element
      },
      y: restingY
    }
  );
}

function createParallax(element, isCompact) {
  const speed = numberAttribute(element, "speed", 10);
  const amount = speed * (isCompact ? 0.48 : 1);
  const trigger = element.closest("section") || element.parentElement || element;

  gsap.to(element, {
    ease: "none",
    scrollTrigger: {
      end: "bottom top",
      invalidateOnRefresh: true,
      scrub: isCompact ? 0.45 : 0.8,
      start: "top bottom",
      trigger
    },
    yPercent: amount
  });
}

function createFormatSequence(container, isCompact) {
  const steps = Array.from(
    container.querySelectorAll("[data-format-step]")
  ).filter(isRenderable);
  const restingY = steps.map(
    (step) => Number(gsap.getProperty(step, "y")) || 0
  );

  if (!steps.length) return;

  gsap.fromTo(
    steps,
    {
      opacity: 0,
      y: (index) => restingY[index] + 42 + index * 15
    },
    {
      duration: 0.95,
      ease: "power3.out",
      opacity: 1,
      scrollTrigger: {
        once: true,
        start: "top 78%",
        trigger: container
      },
      stagger: 0.16,
      y: (index) => restingY[index]
    }
  );

  steps.forEach((step, index) => {
    const depth = [5, -8, 7][index % 3] * (isCompact ? 0.45 : 1);
    gsap.to(step, {
      ease: "none",
      scrollTrigger: {
        end: "bottom top",
        invalidateOnRefresh: true,
        scrub: isCompact ? 0.45 : 0.8,
        start: "top bottom",
        trigger: step
      },
      yPercent: depth
    });
  });
}

function createScheduleProgress(element, isCompact) {
  const rail = element.closest("[data-schedule-rail]") || element.parentElement;

  if (!rail) return;

  gsap.fromTo(
    element,
    {
      scaleY: 0,
      transformOrigin: "top center"
    },
    {
      ease: "none",
      scaleY: 1,
      scrollTrigger: {
        end: "bottom 55%",
        invalidateOnRefresh: true,
        scrub: isCompact ? 0.35 : 0.65,
        start: "top 68%",
        trigger: rail
      }
    }
  );
}

function createMagnetic(element, signal) {
  const restingX = Number(gsap.getProperty(element, "x")) || 0;
  const restingY = Number(gsap.getProperty(element, "y")) || 0;
  const moveX = gsap.quickTo(element, "x", {
    duration: 0.42,
    ease: "power3.out"
  });
  const moveY = gsap.quickTo(element, "y", {
    duration: 0.42,
    ease: "power3.out"
  });
  const strength = numberAttribute(element, "magneticStrength", 0.16);

  element.addEventListener("pointermove", (event) => {
    const bounds = element.getBoundingClientRect();
    moveX(
      restingX +
      (event.clientX - bounds.left - bounds.width / 2) * strength
    );
    moveY(
      restingY +
      (event.clientY - bounds.top - bounds.height / 2) * strength
    );
  }, { signal });

  element.addEventListener("pointerleave", () => {
    moveX(restingX);
    moveY(restingY);
  }, { signal });
}

function createTilt(element, signal) {
  const surface = element.closest("[data-hero]") || element;
  const restingRotationX = Number(gsap.getProperty(element, "rotationX")) || 0;
  const restingRotationY = Number(gsap.getProperty(element, "rotationY")) || 0;
  const rotateX = gsap.quickTo(element, "rotationX", {
    duration: 0.55,
    ease: "power3.out"
  });
  const rotateY = gsap.quickTo(element, "rotationY", {
    duration: 0.55,
    ease: "power3.out"
  });
  const scale = gsap.quickTo(element, "scale", {
    duration: 0.55,
    ease: "power3.out"
  });

  gsap.set(element, { transformPerspective: 900 });

  surface.addEventListener("pointermove", (event) => {
    const bounds = surface.getBoundingClientRect();
    const normalizedX = (event.clientX - bounds.left) / bounds.width - 0.5;
    const normalizedY = (event.clientY - bounds.top) / bounds.height - 0.5;
    rotateX(restingRotationX + normalizedY * -6);
    rotateY(restingRotationY + normalizedX * 6);
    scale(1.015);
  }, { signal });

  surface.addEventListener("pointerleave", () => {
    rotateX(restingRotationX);
    rotateY(restingRotationY);
    scale(1);
  }, { signal });
}

function initHeaderState() {
  const header = document.querySelector("[data-site-header]");
  let frame = 0;

  if (!header) return () => {};

  const update = () => {
    frame = 0;
    header.classList.toggle("is-scrolled", window.scrollY > 24);
  };
  const requestUpdate = () => {
    if (!frame) frame = requestAnimationFrame(update);
  };

  window.addEventListener("scroll", requestUpdate, { passive: true });
  update();

  return () => {
    window.removeEventListener("scroll", requestUpdate);
    cancelAnimationFrame(frame);
    header.classList.remove("is-scrolled");
  };
}

export function initTransitions(root = document.documentElement) {
  const cleanupHeaderState = initHeaderState();
  const media = gsap.matchMedia(root);

  media.add(MOTION_QUERY, (context) => {
    const controller = new AbortController();
    const finePointer = window.matchMedia(FINE_POINTER_QUERY);
    const compactViewport = window.matchMedia("(max-width: 899px)");
    const initialized = {
      format: new WeakSet(),
      hero: new WeakSet(),
      magnetic: new WeakSet(),
      parallax: new WeakSet(),
      progress: new WeakSet(),
      reveal: new WeakSet(),
      tilt: new WeakSet()
    };
    const pendingRoots = new Set();
    let alive = true;
    let refreshTimer = 0;
    let scanFrame = 0;

    const scheduleRefresh = () => {
      window.clearTimeout(refreshTimer);
      refreshTimer = window.setTimeout(() => {
        if (alive) ScrollTrigger.refresh();
      }, 100);
    };

    context.add("scanMotionNodes", (scanRoot = document) => {
      findWithin(scanRoot, "[data-hero]").forEach((hero) => {
        if (initialized.hero.has(hero) || !isRenderable(hero)) return;
        initialized.hero.add(hero);
        createHeroIntro(hero);
      });

      const formatContainers = new Set();
      findWithin(scanRoot, "[data-format-step]").forEach((step) => {
        const container = step.closest(".format-timeline") || step.parentElement;
        if (container) formatContainers.add(container);
      });
      findWithin(scanRoot, ".format-timeline").forEach((container) => {
        formatContainers.add(container);
      });
      formatContainers.forEach((container) => {
        if (initialized.format.has(container) || !isRenderable(container)) return;
        initialized.format.add(container);
        createFormatSequence(container, compactViewport.matches);
      });

      findWithin(scanRoot, "[data-reveal]").forEach((element) => {
        if (initialized.reveal.has(element) || !isRenderable(element)) return;
        initialized.reveal.add(element);
        createReveal(element);
      });

      findWithin(scanRoot, "[data-parallax]").forEach((element) => {
        if (initialized.parallax.has(element) || !isRenderable(element)) return;
        initialized.parallax.add(element);
        createParallax(element, compactViewport.matches);
      });

      findWithin(scanRoot, "[data-schedule-progress]").forEach((element) => {
        if (initialized.progress.has(element) || !isRenderable(element)) return;
        initialized.progress.add(element);
        createScheduleProgress(element, compactViewport.matches);
      });

      if (finePointer.matches) {
        findWithin(scanRoot, "[data-magnetic]").forEach((element) => {
          if (initialized.magnetic.has(element)) return;
          initialized.magnetic.add(element);
          createMagnetic(element, controller.signal);
        });

        findWithin(scanRoot, "[data-tilt]").forEach((element) => {
          if (initialized.tilt.has(element)) return;
          initialized.tilt.add(element);
          createTilt(element, controller.signal);
        });
      }
    });

    const scheduleScan = (scanRoot = document) => {
      if (!alive) return;
      pendingRoots.add(scanRoot);
      if (scanFrame) return;

      scanFrame = requestAnimationFrame(() => {
        scanFrame = 0;
        if (!alive) return;
        pendingRoots.forEach((pendingRoot) => {
          context.scanMotionNodes(pendingRoot);
        });
        pendingRoots.clear();
        scheduleRefresh();
      });
    };

    const mutationObserver = new MutationObserver((records) => {
      records.forEach((record) => {
        record.addedNodes.forEach((node) => {
          if (node.nodeType === Node.ELEMENT_NODE) scheduleScan(node);
        });
      });
    });
    mutationObserver.observe(document.body, {
      childList: true,
      subtree: true
    });

    const resizeObserver = new ResizeObserver(() => {
      scheduleScan(document);
    });
    resizeObserver.observe(document.body);
    const main = document.querySelector("main");
    if (main) resizeObserver.observe(main);

    const requestRefresh = (event) => {
      const requestedRoot = event.detail?.root;
      scheduleScan(
        requestedRoot && typeof requestedRoot.querySelectorAll === "function"
          ? requestedRoot
          : document
      );
    };
    window.addEventListener(MOTION_REFRESH_EVENT, requestRefresh, {
      signal: controller.signal
    });
    finePointer.addEventListener("change", () => scheduleScan(document), {
      signal: controller.signal
    });
    window.addEventListener("load", scheduleRefresh, {
      once: true,
      signal: controller.signal
    });

    context.scanMotionNodes(document);
    scheduleRefresh();

    if (document.fonts?.ready) {
      document.fonts.ready.then(() => {
        if (alive) scheduleRefresh();
      });
    }

    return () => {
      alive = false;
      controller.abort();
      mutationObserver.disconnect();
      resizeObserver.disconnect();
      cancelAnimationFrame(scanFrame);
      window.clearTimeout(refreshTimer);
      pendingRoots.clear();
    };
  });

  return () => {
    media.revert();
    cleanupHeaderState();
  };
}
