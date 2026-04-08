import { cn } from "@/lib/cn";
import type { ReactNode } from "react";

interface LabelTagProps {
  children: ReactNode;
  className?: string;
}

export function LabelTag({ children, className }: LabelTagProps) {
  return (
    <span
      className={cn(
        "rounded bg-accent/10 px-1.5 py-0.5 text-xs text-accent-light",
        className,
      )}
    >
      {children}
    </span>
  );
}
