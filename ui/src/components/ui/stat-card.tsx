import { cn } from "@/lib/cn";
import { Card, CardContent } from "@/components/ui/card";
import type { ReactNode } from "react";
import { Link } from "react-router-dom";

interface StatCardProps {
  label: string;
  value: ReactNode;
  detail?: ReactNode;
  truncateDetail?: boolean;
  className?: string;
  // to makes the card a link to the page that shows what it counts, such as
  // the Jobs page filtered on failed.
  to?: string;
}

export function StatCard({
  label,
  value,
  detail,
  truncateDetail,
  className,
  to,
}: StatCardProps) {
  const card = (
    <Card
      className={cn(
        "p-3",
        to && "transition-colors group-hover:border-primary",
        className,
      )}
    >
      <CardContent className="p-0">
        <p className="text-xs font-medium text-text-muted">{label}</p>
        <p className="text-lg font-semibold text-text">{value}</p>
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
  if (!to) {
    return card;
  }

  return (
    <Link to={to} className="group block rounded-lg focus-visible:outline-2">
      {card}
    </Link>
  );
}
