import { cn } from "@/lib/cn";
import type { ElementType, ReactNode } from "react";

interface MetricValueProps {
  icon: ElementType;
  children: ReactNode;
  className?: string;
}

export function MetricValue({
  icon: Icon,
  children,
  className,
}: MetricValueProps) {
  return (
    <div
      className={cn(
        "flex items-center gap-1.5 text-xs text-text-muted",
        className,
      )}
    >
      <Icon className="h-3 w-3" />
      {children}
    </div>
  );
}
