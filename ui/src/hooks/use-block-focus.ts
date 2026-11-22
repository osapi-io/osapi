import { useCallback, type RefObject } from "react";

/**
 * Returns a callback that focuses a block, after a short delay so React has
 * rendered it: its first input, or the card itself for a block with no
 * input, such as a List. Focus brings the block into view, so a person sees
 * the block that they just added, however long the stack is.
 */
export function useBlockFocus(
  blockRefs: RefObject<Map<string, HTMLElement> | null>,
) {
  return useCallback(
    (id: string) => {
      setTimeout(() => {
        const el = blockRefs.current?.get(id);
        if (!el) return;
        el.scrollIntoView({ behavior: "smooth", block: "center" });
        const input = el.querySelector(
          "input, textarea, select",
        ) as HTMLElement | null;
        if (input) {
          input.focus({ preventScroll: true });
        } else {
          (document.activeElement as HTMLElement)?.blur();
          el.setAttribute("tabindex", "-1");
          el.focus({ preventScroll: true });
        }
      }, 50);
    },
    [blockRefs],
  );
}
