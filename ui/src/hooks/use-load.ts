import { useCallback, useEffect, useState } from "react";

/**
 * useLoad loads a value once on mount, and again on refresh. A load that
 * fails leaves the value as it was: a picker or a suggestion list stays
 * empty rather than breaking the form it is in. A page that must show a
 * failure fetches with its own error state instead. load is a module
 * function in each caller, so it is read once.
 */
export function useLoad<T>(load: () => Promise<T | undefined>, initial: T) {
  const [value, setValue] = useState<T>(initial);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;
    const fetch = async () => {
      try {
        const v = await load();
        if (mounted && v !== undefined) setValue(v);
      } catch {
        // The value stays as it was.
      }
      if (mounted) setLoading(false);
    };
    fetch();
    return () => {
      mounted = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const refresh = useCallback(async () => {
    try {
      const v = await load();
      if (v !== undefined) setValue(v);
    } catch {
      // The value stays as it was.
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return { value, loading, refresh };
}
