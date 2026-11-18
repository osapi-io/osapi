import { cn } from "@/lib/cn";
import { AlertTriangle } from "lucide-react";

interface ConditionAlertProps {
  type: string;
  className?: string;
}

export function ConditionAlert({ type, className }: ConditionAlertProps) {
  return (
    <div
      className={cn(
        "flex items-center gap-1.5 rounded bg-status-pending/10 px-2 py-1 text-xs text-status-pending",
        className,
      )}
    >
      <AlertTriangle className="h-3 w-3" />
      {type}
    </div>
  );
}
