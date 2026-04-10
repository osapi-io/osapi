import { useEffect, type RefObject } from "react";

/**
 * Calls `callback` when a mousedown event occurs outside of `ref`.
 * Only active when `enabled` is true (defaults to true).
 */
export function useOutsideClick(
  ref: RefObject<HTMLElement | null>,
  callback: () => void,
  enabled = true,
) {
  useEffect(() => {
    if (!enabled) return;
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        callback();
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [ref, callback, enabled]);
}
