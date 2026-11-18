import { Loader2 } from "lucide-react";
import { cn } from "@/lib/cn";
import { spinnerColor } from "@/lib/spinner-color";

interface SpinnerProps {
  /** page centers a large spinner in the content; inline is a small one. */
  size?: "page" | "inline";
  className?: string;
}

/**
 * Spinner is the loading state of a page or of a part of one, in the one
 * heading color that spinner-color selects for the page load.
 */
export function Spinner({ size = "page", className }: SpinnerProps) {
  if (size === "inline") {
    return (
      <Loader2
        className={cn("h-5 w-5 animate-spin", spinnerColor, className)}
      />
    );
  }

  return (
    <div className={cn("flex items-center justify-center py-20", className)}>
      <Loader2 className={cn("h-8 w-8 animate-spin", spinnerColor)} />
    </div>
  );
}
