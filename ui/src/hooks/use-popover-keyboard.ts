import { useState, useCallback, type RefObject } from "react";

interface UsePopoverKeyboardOptions {
  /** Total number of items in the list */
  itemCount: number;
  /** Called when an item is selected via Enter */
  onSelect: (index: number) => void;
  /** Called to close the popover */
  onClose: () => void;
  /** Called to open the popover */
  onOpen?: () => void;
  /** Whether the popover is open */
  open: boolean;
  /** Ref to the list container for scroll-into-view */
  listRef?: RefObject<HTMLDivElement | null>;
}

/**
 * Keyboard navigation for popover-based dropdowns and pickers.
 * Returns activeIndex, setActiveIndex, and a keyDown handler.
 *
 * Supports Arrow Up/Down to navigate, Enter to select, Escape to close.
 * When closed, Arrow Down or Enter opens the popover.
 */
export function usePopoverKeyboard({
  itemCount,
  onSelect,
  onClose,
  onOpen,
  open,
  listRef,
}: UsePopoverKeyboardOptions) {
  const [activeIndex, setActiveIndex] = useState(-1);

  const reset = useCallback(() => setActiveIndex(-1), []);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (!open) {
        if (e.key === "ArrowDown" || e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onOpen?.();
          setActiveIndex(0);
        }
        return;
      }

      if (e.key === "Escape") {
        e.preventDefault();
        onClose();
        setActiveIndex(-1);
      } else if (e.key === "ArrowDown") {
        e.preventDefault();
        const next = activeIndex < itemCount - 1 ? activeIndex + 1 : 0;
        setActiveIndex(next);
        const buttons = listRef?.current?.querySelectorAll("button");
        buttons?.[next]?.scrollIntoView({ block: "nearest" });
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        const next = activeIndex > 0 ? activeIndex - 1 : itemCount - 1;
        setActiveIndex(next);
        const buttons = listRef?.current?.querySelectorAll("button");
        buttons?.[next]?.scrollIntoView({ block: "nearest" });
      } else if (e.key === "Enter" && activeIndex >= 0) {
        e.preventDefault();
        onSelect(activeIndex);
        setActiveIndex(-1);
      }
    },
    [open, activeIndex, itemCount, onSelect, onClose, onOpen, listRef],
  );

  return {
    activeIndex,
    setActiveIndex,
    resetActiveIndex: reset,
    handleKeyDown,
  };
}
