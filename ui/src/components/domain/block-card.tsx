import { forwardRef, type ReactNode } from "react";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Text } from "@/components/ui/text";
import { cn } from "@/lib/cn";
import type { BlockStatus } from "@/hooks/use-stack";
import { Check, X, Loader2, XCircle } from "lucide-react";
import { TargetPicker } from "@/components/domain/target-picker";
import { IconButton } from "@/components/ui/icon-button";
import { BlockKindIcon } from "@/components/domain/block-kind";
import { blockKindLabels, type BlockKind } from "@/lib/permissions";

interface BlockCardProps {
  label: string;
  description: string;
  /** Whether the block reads the host, writes it or runs an action on it. */
  kind?: BlockKind;
  status: BlockStatus;
  focused?: boolean;
  error?: string;
  target?: string;
  onTargetChange?: (target: string) => void;
  onRemove: () => void;
  onFocusCard?: () => void;
  children?: ReactNode;
}

const statusToBadge: Record<
  BlockStatus,
  {
    variant: "ready" | "pending" | "running" | "error" | "applied" | "muted";
    label: string;
  }
> = {
  unchecked: { variant: "muted", label: "" },
  pending: { variant: "pending", label: "Pending" },
  ready: { variant: "ready", label: "Ready" },
  applying: { variant: "pending", label: "Applying" },
  applied: { variant: "applied", label: "Applied" },
  error: { variant: "error", label: "Error" },
};

const statusToCardVariant: Record<
  BlockStatus,
  "default" | "active" | "pending" | "applied" | "error"
> = {
  unchecked: "default",
  pending: "pending",
  ready: "active",
  applying: "pending",
  applied: "applied",
  error: "error",
};

function StatusIcon({ status }: { status: BlockStatus }) {
  switch (status) {
    case "applying":
      return <Loader2 className="h-4 w-4 animate-spin text-status-pending" />;
    case "applied":
      return <Check className="h-4 w-4 text-status-applied" />;
    case "error":
      return <X className="h-4 w-4 text-status-error" />;
    default:
      return null;
  }
}

export const BlockCard = forwardRef<HTMLDivElement, BlockCardProps>(
  (
    {
      label,
      description,
      kind,
      status,
      focused,
      error,
      target = "_all",
      onTargetChange,
      onRemove,
      onFocusCard,
      children,
    },
    ref,
  ) => {
    const badge = statusToBadge[status];

    return (
      <div ref={ref} onClick={onFocusCard} className="outline-none">
        <Card
          variant={statusToCardVariant[status]}
          className={cn(
            // The block being edited has the border that a card under the
            // pointer has, not a blue ring. A status keeps its own border.
            focused &&
              (status === "unchecked" || status === "ready") &&
              "border-white/20",
            status === "applying" && "ring-1 ring-status-pending/40",
          )}
        >
          <CardHeader>
            <div className="flex-1">
              <CardTitle>{label}</CardTitle>
              <Text variant="muted" as="p">
                {description}
              </Text>
            </div>
            <div className="flex items-center gap-2">
              {kind && (
                <Badge variant="muted">
                  <BlockKindIcon kind={kind} className="text-current" />
                  {blockKindLabels[kind]}
                </Badge>
              )}
              <StatusIcon status={status} />
              {badge.label && (
                <Badge variant={badge.variant}>{badge.label}</Badge>
              )}
              <IconButton
                icon={XCircle}
                variant="danger"
                onClick={onRemove}
                title="Remove block"
              />
            </div>
          </CardHeader>
          {(children != null || onTargetChange) && (
            <div className="border-t border-border pt-4">
              {onTargetChange && (
                <div className="mb-3">
                  <TargetPicker value={target} onChange={onTargetChange} />
                  {target === "" && (
                    <Text variant="muted" size="xs" as="p" className="mt-1">
                      Pick the hosts that it changes. It does not apply until it
                      has a target.
                    </Text>
                  )}
                </div>
              )}
              {children}
              {error && (
                <Text variant="error" as="p" className="mt-2">
                  {error}
                </Text>
              )}
            </div>
          )}
        </Card>
      </div>
    );
  },
);
BlockCard.displayName = "BlockCard";
