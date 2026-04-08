import { cn } from "@/lib/cn";
import { Text } from "@/components/ui/text";
import type { ReactNode } from "react";

interface EmptyStateProps {
  message: string;
  icon?: ReactNode;
  className?: string;
}

export function EmptyState({ message, icon, className }: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed border-border py-20",
        className,
      )}
    >
      {icon}
      <Text variant="muted" size="sm">
        {message}
      </Text>
    </div>
  );
}
