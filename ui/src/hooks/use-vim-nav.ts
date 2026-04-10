import { useEffect, useRef } from "react";
import { features } from "@/lib/features";

interface VimNavItem {
  id: string;
}

interface VimNavOptions<T extends VimNavItem> {
  /** List of navigable items */
  items: T[];
  /** Currently focused item ID */
  focusedId: string | null;
  /** Set the focused item */
  setFocusedId: (id: string | null) => void;
  /** Called when dd is pressed on a focused item */
  onDelete?: (id: string) => void;
  /** Called on Ctrl/Cmd+Enter */
  onExecute?: () => void;
  /** Whether execution is allowed */
  canExecute?: boolean;
  /** Whether execution is in progress */
  executing?: boolean;
  /** Ref map for scrolling items into view */
  itemRefs?: React.RefObject<Map<string, HTMLElement> | null>;
}

export function useVimNav<T extends VimNavItem>({
  items,
  focusedId,
  setFocusedId,
  onDelete,
  onExecute,
  canExecute = false,
  executing = false,
  itemRefs,
}: VimNavOptions<T>) {
  const lastKeyRef = useRef<string | null>(null);

  useEffect(() => {
    if (!features.keyboard) return;
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement).tagName;
      const inInput = tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT";

      // Ctrl/Cmd+Enter to execute — works even in inputs
      if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
        e.preventDefault();
        if (canExecute && !executing && onExecute) onExecute();
        return;
      }

      // Escape in an input — blur and set focus to the containing item
      if (inInput && e.key === "Escape") {
        const target = e.target as HTMLElement;
        target.blur();
        // Find which item contains this input
        const refs = itemRefs?.current;
        if (refs) {
          const entries = Array.from(refs.entries());
          for (let i = 0; i < entries.length; i++) {
            if (entries[i][1].contains(target)) {
              setFocusedId(entries[i][0]);
              break;
            }
          }
        }
        lastKeyRef.current = null;
        return;
      }

      // Skip vim keys when typing in an input
      if (inInput) {
        lastKeyRef.current = null;
        return;
      }

      // Escape — clear focus
      if (e.key === "Escape") {
        setFocusedId(null);
        lastKeyRef.current = null;
        return;
      }

      // j/k — navigate items
      if ((e.key === "j" || e.key === "k") && items.length > 0) {
        e.preventDefault();
        const currentIdx = focusedId
          ? items.findIndex((item) => item.id === focusedId)
          : -1;

        let nextIdx: number;
        if (e.key === "j") {
          nextIdx = currentIdx < items.length - 1 ? currentIdx + 1 : 0;
        } else {
          nextIdx = currentIdx > 0 ? currentIdx - 1 : items.length - 1;
        }

        const nextId = items[nextIdx].id;
        setFocusedId(nextId);
        itemRefs?.current
          ?.get(nextId)
          ?.scrollIntoView({ behavior: "smooth", block: "nearest" });
        lastKeyRef.current = null;
        return;
      }

      // dd — delete focused item
      if (e.key === "d" && onDelete) {
        if (lastKeyRef.current === "d" && focusedId) {
          e.preventDefault();
          const currentIdx = items.findIndex((item) => item.id === focusedId);
          // Move focus before deleting
          if (items.length <= 1) {
            setFocusedId(null);
          } else if (currentIdx < items.length - 1) {
            setFocusedId(items[currentIdx + 1].id);
          } else {
            setFocusedId(items[currentIdx - 1].id);
          }
          onDelete(focusedId);
          lastKeyRef.current = null;
          return;
        }
        lastKeyRef.current = "d";
        setTimeout(() => {
          lastKeyRef.current = null;
        }, 500);
        return;
      }

      lastKeyRef.current = null;
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [
    items,
    focusedId,
    setFocusedId,
    onDelete,
    onExecute,
    canExecute,
    executing,
    itemRefs,
  ]);
}
