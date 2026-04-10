import { useEffect } from "react";
import { features } from "@/lib/features";

/**
 * Vim-style page scrolling with j/k when not in an input.
 * Scrolls by a fixed step (default 80px) per keypress.
 */
export function useVimScroll(step = 80) {
  useEffect(() => {
    if (!features.keyboard) return;
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement).tagName;
      const inInput = tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT";
      if (inInput) return;

      if (e.key === "j") {
        e.preventDefault();
        window.scrollBy({ top: step, behavior: "smooth" });
      } else if (e.key === "k") {
        e.preventDefault();
        window.scrollBy({ top: -step, behavior: "smooth" });
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [step]);
}
