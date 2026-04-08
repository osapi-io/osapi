import { cn } from "@/lib/cn";
import { HealthDot } from "@/components/ui/health-dot";
import { Text } from "@/components/ui/text";

interface ComponentRowProps {
  name: string;
  status: string;
  address?: string;
  isLast?: boolean;
}

function isHealthy(status: string) {
  return status === "ok" || status === "healthy" || status === "ready";
}

export function ComponentRow({
  name,
  status,
  address,
  isLast,
}: ComponentRowProps) {
  const ok = isHealthy(status);

  return (
    <div
      className={cn(
        "flex items-center gap-2 px-3 py-1.5 text-xs",
        !isLast && "border-b border-border/20",
      )}
    >
      <HealthDot status={ok} />
      <Text className="font-medium">{name}</Text>
      {address && <Text variant="muted">{address}</Text>}
      <Text variant={ok ? "primary" : "error"} className="ml-auto">
        {status}
      </Text>
    </div>
  );
}
