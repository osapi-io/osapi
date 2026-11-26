import { cn } from "@/lib/cn";
import type { ReactNode } from "react";

interface PageHeaderProps {
  title: string;
  subtitle?: ReactNode;
  actions?: ReactNode;
  className?: string;
}

export function PageHeader({
  title,
  subtitle,
  actions,
  className,
}: PageHeaderProps) {
  return (
    <div className={cn("mb-6 flex items-center justify-between", className)}>
      <div>
        <h1 className="text-2xl font-bold text-white">{title}</h1>
        {subtitle && (
          <div className="mt-1 text-sm text-text-muted">{subtitle}</div>
        )}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  );
}
