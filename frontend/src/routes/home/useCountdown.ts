import { useEffect, useState } from "react";

export interface CountdownState {
  /** Milliseconds left, clamped at 0; null when the deadline is unknown. */
  remainingMs: number | null;
  /** True once the deadline is reached — the terminal state. */
  finished: boolean;
}

function pad(value: number): string {
  return String(value).padStart(2, "0");
}

/** «Nd HH:MM:SS» — e.g. «3д 04:12:45». */
export function formatCountdown(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000);
  const days = Math.floor(totalSeconds / 86400);
  const hours = Math.floor((totalSeconds % 86400) / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  return `${days}д ${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
}

/*
 * Live countdown to the registration deadline. Ticks every second, switches
 * to the terminal state at/after the deadline and stops its own interval
 * (LIFECYCLE-001/003): the interval is cleared both on teardown and as soon
 * as the terminal state is reached.
 */
export function useCountdown(deadline: string | undefined): CountdownState {
  const target = deadline ? new Date(deadline).getTime() : Number.NaN;
  const valid = deadline !== undefined && !Number.isNaN(target);
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    if (!valid) return;
    setNow(Date.now());
    const id = setInterval(() => {
      const current = Date.now();
      setNow(current);
      if (target - current <= 0) clearInterval(id);
    }, 1000);
    return () => clearInterval(id);
  }, [valid, target]);

  if (!valid) return { remainingMs: null, finished: false };
  const remaining = Math.max(0, target - now);
  return { remainingMs: remaining, finished: remaining <= 0 };
}
