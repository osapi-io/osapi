import { StatusIcon } from "@/components/ui/status-icon";
import { Text } from "@/components/ui/text";

interface HostGroupHeaderProps {
  hostname: string;
  status: "ok" | "failed" | "skipped";
  detail?: string;
}

export function HostGroupHeader({
  hostname,
  status,
  detail,
}: HostGroupHeaderProps) {
  return (
    <div className="flex items-center gap-2 bg-card/50 px-4 py-2">
      <StatusIcon status={status} />
      <Text className="font-semibold">{hostname}</Text>
      {detail && <Text variant="muted">{detail}</Text>}
    </div>
  );
}
