import { cn } from "@/lib/cn";
import { HealthDot } from "@/components/ui/health-dot";
import { Text } from "@/components/ui/text";

interface ComponentRowProps {
  name: string;
  status: string;
  address?: string;
  // Where a person opens the component, from its url config key. An http or
  // https url is a link; another scheme, such as nats://, is text.
  url?: string;
  isLast?: boolean;
}

function isWebURL(url: string) {
  return /^https?:\/\//i.test(url);
}

function isHealthy(status: string) {
  return status === "ok" || status === "healthy" || status === "ready";
}

export function ComponentRow({
  name,
  status,
  address,
  url,
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
      {url &&
        (isWebURL(url) ? (
          <a
            href={url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary hover:underline"
          >
            {url}
          </a>
        ) : (
          <Text>{url}</Text>
        ))}
      <Text variant={ok ? "primary" : "error"} className="ml-auto">
        {status}
      </Text>
    </div>
  );
}
