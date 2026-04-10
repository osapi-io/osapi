import { cn } from "@/lib/cn";
import type { ReactNode } from "react";

interface KbdProps {
  children: ReactNode;
  className?: string;
}

export function Kbd({ children, className }: KbdProps) {
  return (
    <kbd
      className={cn(
        "rounded border border-white/10 bg-white/5 px-1 py-0.5 font-mono text-[10px] text-text-muted",
        className,
      )}
    >
      {children}
    </kbd>
  );
}
