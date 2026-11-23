import { Eye, Pencil, Zap } from "lucide-react";
import { cn } from "@/lib/cn";
import { blockKindLabels, type BlockKind } from "@/lib/permissions";

const icons: Record<BlockKind, typeof Eye> = {
  read: Eye,
  write: Pencil,
  action: Zap,
};

/**
 * BlockKindIcon marks a block as a read, a write or an action, on its chip
 * in the panel and on its card.
 */
export function BlockKindIcon({
  kind,
  className,
}: {
  kind: BlockKind;
  className?: string;
}) {
  const Icon = icons[kind];
  return (
    <Icon
      aria-label={blockKindLabels[kind]}
      className={cn("h-3 w-3 shrink-0 text-text-muted", className)}
    />
  );
}
