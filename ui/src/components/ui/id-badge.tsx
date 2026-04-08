import { cn } from "@/lib/cn";
import type { HTMLAttributes } from "react";

interface IdBadgeProps extends HTMLAttributes<HTMLSpanElement> {
  interactive?: boolean;
}

export function IdBadge({
  interactive = false,
  className,
  children,
  ...props
}: IdBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-md border border-accent/20 bg-accent/5 px-2.5 py-1 font-mono text-xs text-accent-light",
        interactive &&
          "cursor-pointer transition-colors hover:border-accent/40 hover:bg-accent/15",
        className,
      )}
      {...props}
    >
      {children}
    </span>
  );
}
