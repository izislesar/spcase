const DIALOG_SELECTOR = '[role="dialog"][aria-modal="true"]';
const MENU_SELECTOR = "details.mobile-menu[open]";
const BACKGROUND_EXEMPT_SELECTOR = [
  "[aria-live]",
  "[role='alert']",
  "[data-focus-exempt]"
].join(", ");
const FOCUSABLE_SELECTOR = [
  "a[href]",
  "area[href]",
  "button:not([disabled])",
  "input:not([disabled]):not([type='hidden'])",
  "select:not([disabled])",
  "textarea:not([disabled])",
  "summary",
  "[contenteditable='true']",
  "[tabindex]:not([tabindex='-1'])"
].join(", ");

function isVisible(element) {
  return (
    element instanceof HTMLElement &&
    element.isConnected &&
    element.getClientRects().length > 0 &&
    window.getComputedStyle(element).visibility !== "hidden"
  );
}

function focusableWithin(scope) {
  return Array.from(scope.querySelectorAll(FOCUSABLE_SELECTOR)).filter((element) => (
    isVisible(element) &&
    !element.closest("[inert]") &&
    !element.closest("[aria-hidden='true']")
  ));
}

function focusElement(element) {
  if (!(element instanceof HTMLElement) || !element.isConnected) return;
  try {
    element.focus({ preventScroll: true });
  } catch {
    element.focus();
  }
}

function makeBackgroundInert(scope) {
  const previousStates = new Map();
  let branch = scope;

  while (branch.parentElement && branch.parentElement !== document.documentElement) {
    const parent = branch.parentElement;
    for (const sibling of parent.children) {
      if (
        sibling === branch ||
        !(sibling instanceof HTMLElement) ||
        sibling.matches(BACKGROUND_EXEMPT_SELECTOR)
      ) {
        continue;
      }
      if (!previousStates.has(sibling)) {
        previousStates.set(sibling, sibling.hasAttribute("inert"));
        sibling.setAttribute("inert", "");
      }
    }
    branch = parent;
  }

  return () => {
    for (const [element, wasInert] of previousStates) {
      if (!element.isConnected || wasInert) continue;
      element.removeAttribute("inert");
    }
  };
}

function findActiveScope() {
  const dialogs = Array.from(document.querySelectorAll(DIALOG_SELECTOR)).filter(isVisible);
  if (dialogs.length > 0) {
    return {
      element: dialogs[dialogs.length - 1],
      kind: "dialog"
    };
  }

  const menus = Array.from(document.querySelectorAll(MENU_SELECTOR)).filter(isVisible);
  if (menus.length > 0) {
    return {
      element: menus[menus.length - 1],
      kind: "menu"
    };
  }

  return null;
}

function associatedTrigger(scope) {
  if (!scope.id) return null;
  return Array.from(document.querySelectorAll("[aria-controls]")).find((element) => (
    element.getAttribute("aria-controls")?.split(/\s+/).includes(scope.id)
  )) || null;
}

export function initFocusManagement() {
  const controller = new AbortController();
  let activeScope = null;
  let scanFrame = 0;
  let stopped = false;

  const firstFocusTarget = (scope) => {
    const preferred = scope.element.querySelector("[data-focus-initial]");
    if (preferred && isVisible(preferred) && !preferred.matches(":disabled")) {
      return preferred;
    }
    return focusableWithin(scope.element)[0] || scope.element;
  };

  const deactivate = (restoreFocus = true) => {
    if (!activeScope) return;

    const scope = activeScope;
    activeScope = null;
    scope.restoreBackground();

    if (scope.addedFallbackTabindex) {
      scope.element.removeAttribute("tabindex");
    }

    if (restoreFocus && scope.restoreTarget?.isConnected) {
      requestAnimationFrame(() => {
        if (!activeScope) focusElement(scope.restoreTarget);
      });
    }
  };

  const activate = (nextScope) => {
    const currentFocus = document.activeElement instanceof HTMLElement
      ? document.activeElement
      : null;
    const menuSummary = nextScope.kind === "menu"
      ? nextScope.element.querySelector("summary")
      : null;
    const restoreTarget = nextScope.kind === "menu"
      ? menuSummary
      : (
          currentFocus &&
          currentFocus !== document.body &&
          !nextScope.element.contains(currentFocus)
            ? currentFocus
            : associatedTrigger(nextScope.element)
        );
    const addedFallbackTabindex = !nextScope.element.hasAttribute("tabindex");

    if (addedFallbackTabindex) {
      nextScope.element.setAttribute("tabindex", "-1");
    }

    activeScope = {
      ...nextScope,
      addedFallbackTabindex,
      restoreBackground: makeBackgroundInert(nextScope.element),
      restoreTarget
    };

    requestAnimationFrame(() => {
      if (!activeScope || activeScope.element !== nextScope.element) return;
      if (
        nextScope.kind === "menu" &&
        activeScope.element.contains(document.activeElement)
      ) {
        return;
      }
      focusElement(firstFocusTarget(activeScope));
    });
  };

  const sync = () => {
    scanFrame = 0;
    if (stopped) return;

    const nextScope = findActiveScope();
    if (activeScope?.element === nextScope?.element) return;

    deactivate(!nextScope);
    if (nextScope) activate(nextScope);
  };

  const scheduleSync = () => {
    if (stopped || scanFrame) return;
    scanFrame = requestAnimationFrame(sync);
  };

  const observer = new MutationObserver(scheduleSync);
  observer.observe(document.body, {
    attributeFilter: ["open"],
    attributes: true,
    childList: true,
    subtree: true
  });

  // Alpine's x-show changes only the dialog's inline display style. Observe
  // those few nodes directly so focus activation cannot race the x-effect
  // scroll-lock event, while GSAP style writes elsewhere stay unobserved.
  const dialogObserver = new MutationObserver(scheduleSync);
  document.querySelectorAll(DIALOG_SELECTOR).forEach((dialog) => {
    dialogObserver.observe(dialog, {
      attributeFilter: ["style"],
      attributes: true
    });
  });

  document.addEventListener("keydown", (event) => {
    if (!activeScope || event.key !== "Tab") return;

    const focusable = focusableWithin(activeScope.element);
    if (focusable.length === 0) {
      event.preventDefault();
      focusElement(activeScope.element);
      return;
    }

    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    const current = document.activeElement;
    if (!activeScope.element.contains(current)) {
      event.preventDefault();
      focusElement(event.shiftKey ? last : first);
      return;
    }
    if (event.shiftKey && current === first) {
      event.preventDefault();
      focusElement(last);
      return;
    }
    if (!event.shiftKey && current === last) {
      event.preventDefault();
      focusElement(first);
    }
  }, {
    capture: true,
    signal: controller.signal
  });

  document.addEventListener("focusin", (event) => {
    if (
      activeScope &&
      event.target instanceof Node &&
      !activeScope.element.contains(event.target)
    ) {
      focusElement(firstFocusTarget(activeScope));
    }
  }, {
    capture: true,
    signal: controller.signal
  });

  window.addEventListener("resize", scheduleSync, {
    passive: true,
    signal: controller.signal
  });
  window.addEventListener("spcase:scroll-lock", scheduleSync, {
    signal: controller.signal
  });

  const stop = () => {
    if (stopped) return;
    stopped = true;
    controller.abort();
    observer.disconnect();
    dialogObserver.disconnect();
    cancelAnimationFrame(scanFrame);
    deactivate(false);
  };

  window.addEventListener("pagehide", (event) => {
    if (!event.persisted) stop();
  }, {
    signal: controller.signal
  });

  scheduleSync();
  return stop;
}
