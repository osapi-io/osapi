import { useCallback, type RefObject } from "react";

/**
 * Returns a callback that focuses the first input of the first block
 * in the stack after a short delay (to allow React to render the new block).
 *
 * Used by both the sidebar buttons and the command palette when adding blocks.
 */
export function useBlockFocus(
  blockRefs: RefObject<Map<string, HTMLElement> | null>,
  blockIds: string[],
) {
  return useCallback(() => {
    setTimeout(() => {
      const firstId = blockIds.length > 0 ? blockIds[0] : null;
      const el = firstId
        ? blockRefs.current?.get(firstId)
        : blockRefs.current?.values().next().value;
      if (!el) return;
      const input = el.querySelector(
        "input, textarea, select",
      ) as HTMLElement | null;
      if (input) {
        input.focus();
      } else {
        // No input (e.g. List blocks) — blur sidebar and focus the card
        (document.activeElement as HTMLElement)?.blur();
        el.setAttribute("tabindex", "-1");
        el.focus({ preventScroll: true });
      }
    }, 50);
  }, [blockRefs, blockIds]);
}
