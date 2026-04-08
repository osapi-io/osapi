import { cn } from "@/lib/cn";
import type { ElementType, ReactNode } from "react";

interface SectionLabelProps {
  children: ReactNode;
  icon?: ElementType;
  className?: string;
}

export function SectionLabel({
  children,
  icon: Icon,
  className,
}: SectionLabelProps) {
  return (
    <h2
      className={cn(
        "mb-2 flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-text-muted",
        className,
      )}
    >
      {Icon && <Icon className="h-3.5 w-3.5" />}
      {children}
    </h2>
  );
}
