import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";

/** What a page reads from a call that it repeats. */
export interface Polled<T> {
  data: T | null;
  error: string | null;
  loading: boolean;
  /** Calls again now, as after a write that changed what the call reads. */
  refresh: () => Promise<void>;
}

interface Options {
  /**
   * Whether to call at all. Pass `canCall(operation)`, so a role that the
   * API refuses sends no request that gets a 403.
   */
  enabled?: boolean;
  /** Milliseconds between two calls; 0 calls one time. */
  intervalMs?: number;
}

/** The response of a generated SDK call: its status and its body. */
interface Response {
  status: number;
  data: unknown;
}

/**
 * The body of a response with the status that the caller expects, or an
 * error with the API's message, or with fallback.
 */
export function unwrap<T>(res: Response, fallback: string, ok = 200): T {
  if (res.status === ok) return res.data as T;
  const msg = (res.data as { error?: string } | undefined)?.error;
  throw new Error(msg || fallback);
}

/**
 * Calls fetch at once and then each intervalMs, while enabled. It drops the
 * answer of a call that ends after the page unmounts, or after deps change,
 * so a slow answer never replaces a newer one. A new value of deps calls
 * again at once.
 */
export function usePolling<T>(
  fetch: () => Promise<T>,
  deps: readonly unknown[],
  { enabled = true, intervalMs = 10000 }: Options = {},
): Polled<T> {
  const [data, setData] = useState<T | null>(null);
  const [error, setError] = useState<string | null>(null);
  const generation = useRef(0);
  const fetchRef = useRef(fetch);
  // The latest fetch, with the latest values of the page, for the next call.
  useLayoutEffect(() => {
    fetchRef.current = fetch;
  });

  const run = useCallback(async () => {
    const mine = generation.current;
    try {
      const next = await fetchRef.current();
      if (mine !== generation.current) return;
      setData(next);
      setError(null);
    } catch (e) {
      if (mine !== generation.current) return;
      setError(e instanceof Error ? e.message : String(e));
    }
  }, []);

  useEffect(() => {
    generation.current += 1;
    if (!enabled) return;
    const first = setTimeout(run, 0);
    const id = intervalMs > 0 ? setInterval(run, intervalMs) : undefined;
    return () => {
      generation.current += 1;
      clearTimeout(first);
      if (id) clearInterval(id);
    };
    // deps is the caller's list, as for useEffect.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [enabled, intervalMs, run, ...deps]);

  // Loading until the first answer, as a page shows a spinner only before it
  // has something to show.
  const loading = enabled && data === null && error === null;
  return { data, error, loading, refresh: run };
}
