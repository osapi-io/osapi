import { cn } from "@/lib/cn";

type HealthDotColor = "ok" | "error" | "muted";

interface HealthDotProps {
  status: boolean | HealthDotColor;
  className?: string;
}

const colorMap: Record<HealthDotColor, string> = {
  ok: "bg-primary",
  error: "bg-status-error",
  muted: "bg-text-muted",
};

export function HealthDot({ status, className }: HealthDotProps) {
  const color =
    typeof status === "boolean"
      ? status
        ? colorMap.ok
        : colorMap.error
      : colorMap[status];

  return (
    <div
      className={cn("h-1.5 w-1.5 shrink-0 rounded-full", color, className)}
    />
  );
}
