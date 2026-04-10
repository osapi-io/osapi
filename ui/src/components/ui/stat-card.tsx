import { cn } from "@/lib/cn";
import { Card, CardContent } from "@/components/ui/card";
import type { ReactNode } from "react";

interface StatCardProps {
  label: string;
  value: ReactNode;
  detail?: ReactNode;
  truncateDetail?: boolean;
  className?: string;
}

export function StatCard({
  label,
  value,
  detail,
  truncateDetail,
  className,
}: StatCardProps) {
  return (
    <Card className={cn("p-3", className)}>
      <CardContent className="p-0">
        <p className="text-xs font-semibold uppercase tracking-wider text-text-muted">
          {label}
        </p>
        <p className="text-sm font-bold text-primary">{value}</p>
        {detail != null && (
          <p
            className={cn(
              "text-xs text-text-muted",
              truncateDetail && "truncate",
            )}
          >
            {detail}
          </p>
        )}
      </CardContent>
    </Card>
  );
}
