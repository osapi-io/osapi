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

  // The url sits on its own line under the address, as in the CLI, so a
  // long one is never cut and every row keeps the same shape.
  return (
    <div
      className={cn(
        "px-3 py-1.5 text-xs",
        !isLast && "border-b border-border/20",
      )}
    >
      <div className="flex items-center gap-2">
        <HealthDot status={ok} />
        <Text className="font-medium">{name}</Text>
        {address && <Text variant="muted">{address}</Text>}
        <Text variant={ok ? "primary" : "error"} className="ml-auto">
          {status}
        </Text>
      </div>
      {url && (
        <div className="mt-0.5 flex min-w-0 items-center gap-1 pl-4">
          <Text variant="muted">↳</Text>
          {isWebURL(url) ? (
            <a
              href={url}
              target="_blank"
              rel="noopener noreferrer"
              className="truncate text-text hover:underline"
              title={url}
            >
              {url}
            </a>
          ) : (
            <Text className="truncate" title={url}>
              {url}
            </Text>
          )}
        </div>
      )}
    </div>
  );
}
