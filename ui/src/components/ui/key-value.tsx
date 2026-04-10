import { cn } from "@/lib/cn";
import { Text } from "@/components/ui/text";
import type { ReactNode } from "react";

type KVVariant = "default" | "strong" | "accent" | "mono" | "error";

interface KeyValueProps {
  label: string;
  value: ReactNode;
  variant?: KVVariant;
  className?: string;
}

const valueVariantMap: Record<KVVariant, string> = {
  default: "",
  strong: "font-medium",
  accent: "text-accent-light",
  mono: "font-mono text-text-muted",
  error: "text-status-error",
};

export function KeyValue({
  label,
  value,
  variant = "default",
  className,
}: KeyValueProps) {
  return (
    <div className={cn("flex items-center gap-1.5", className)}>
      <Text variant="muted">{label}:</Text>
      <Text className={valueVariantMap[variant]}>{value}</Text>
    </div>
  );
}
