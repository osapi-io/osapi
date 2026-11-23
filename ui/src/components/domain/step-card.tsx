import { forwardRef, type ReactNode } from "react";
import { Link } from "react-router-dom";
import {
  ArrowDown,
  ArrowUp,
  Check,
  Loader2,
  Server,
  X,
  XCircle,
} from "lucide-react";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Text } from "@/components/ui/text";
import { IconButton } from "@/components/ui/icon-button";
import { cn } from "@/lib/cn";
import { TargetPicker } from "@/components/domain/target-picker";
import { KindMark } from "@/components/domain/method-kind";
import { flatCalls } from "@/lib/calls";
import { methodKindLabels } from "@/lib/catalog";
import type { StepDef } from "@/lib/catalog-index";
import type { StepStatus } from "@/lib/plan";

interface StepCardProps {
  /** The step's place in the plan, from 1. */
  index: number;
  def: StepDef;
  status: StepStatus;
  focused?: boolean;
  error?: string;
  target: string;
  onTargetChange: (target: string) => void;
  onRemove: () => void;
  onMove?: (by: -1 | 1) => void;
  first?: boolean;
  last?: boolean;
  onFocusCard?: () => void;
  children?: ReactNode;
}

const statusBadge: Record<
  StepStatus,
  { variant: "ready" | "pending" | "error" | "applied"; label: string }
> = {
  pending: { variant: "pending", label: "Pending" },
  ready: { variant: "ready", label: "Ready" },
  running: { variant: "pending", label: "Running" },
  done: { variant: "applied", label: "Done" },
  error: { variant: "error", label: "Error" },
};

const cardVariant: Record<
  StepStatus,
  "default" | "active" | "pending" | "applied" | "error"
> = {
  pending: "pending",
  ready: "active",
  running: "pending",
  done: "applied",
  error: "error",
};

function StatusIcon({ status }: { status: StepStatus }) {
  switch (status) {
    case "running":
      return <Loader2 className="h-4 w-4 animate-spin text-status-pending" />;
    case "done":
      return <Check className="h-4 w-4 text-status-applied" />;
    case "error":
      return <X className="h-4 w-4 text-status-error" />;
    default:
      return null;
  }
}

/**
 * StepCard is one step of a plan: its place, the behavior and the method
 * that it runs ("DNS · Update"), the method's signature and the behaviors
 * that it calls, its target, and its form. A write or an action has no
 * target at first; the card asks for the hosts, and the step is pending
 * until it has them. A call of the controller has no target.
 */
export const StepCard = forwardRef<HTMLDivElement, StepCardProps>(
  (
    {
      index,
      def,
      status,
      focused,
      error,
      target,
      onTargetChange,
      onRemove,
      onMove,
      first,
      last,
      onFocusCard,
      children,
    },
    ref,
  ) => {
    const badge = statusBadge[status];
    const calls = def.method ? flatCalls(def.method.calls) : [];

    return (
      <div ref={ref} onClick={onFocusCard} className="outline-none">
        <Card
          variant={cardVariant[status]}
          className={cn(
            focused &&
              (status === "ready" || status === "pending") &&
              "border-white/20",
            status === "running" && "ring-1 ring-status-pending/40",
          )}
        >
          <div className="flex items-start gap-3 pb-3">
            <span className="mt-0.5 w-5 shrink-0 text-right font-mono text-xs text-text-muted">
              {index}
            </span>
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-x-2">
                <h3 className="text-sm font-semibold text-white">
                  {def.title}
                  <span className="px-1.5 text-text-muted">·</span>
                  {def.label}
                </h3>
                <span className="font-mono text-xs text-text-muted">
                  {def.run}
                </span>
              </div>
              <Text variant="muted" as="p" className="mt-0.5">
                {def.doc}
              </Text>
              {def.method && (
                <p className="mt-1.5 break-words font-mono text-[11px] text-text-muted/80">
                  {def.method.input} → {def.method.output}
                </p>
              )}
              {calls.length > 0 && (
                <p className="mt-0.5 font-mono text-[11px] text-text-muted/80">
                  calls{" "}
                  {calls.map((c, i) => (
                    <span key={c}>
                      {i > 0 && ", "}
                      <Link
                        to={`/behaviors/${c.split(".")[0]}#${c.split(".")[1]}`}
                        className="hover:text-heading-1"
                        onClick={(e) => e.stopPropagation()}
                      >
                        {c}
                      </Link>
                    </span>
                  ))}
                </p>
              )}
            </div>
            <div className="flex items-center gap-1.5">
              <Badge variant="muted">
                <KindMark kind={def.kind} className="border-current" />
                {methodKindLabels[def.kind]}
              </Badge>
              <StatusIcon status={status} />
              <Badge variant={badge.variant}>{badge.label}</Badge>
              {onMove && (
                <>
                  <IconButton
                    icon={ArrowUp}
                    onClick={() => onMove(-1)}
                    disabled={first}
                    title="Move the step up"
                    className="disabled:opacity-30"
                  />
                  <IconButton
                    icon={ArrowDown}
                    onClick={() => onMove(1)}
                    disabled={last}
                    title="Move the step down"
                    className="disabled:opacity-30"
                  />
                </>
              )}
              <IconButton
                icon={XCircle}
                variant="danger"
                onClick={onRemove}
                title="Remove the step"
              />
            </div>
          </div>
          <div className="space-y-3 border-t border-border pt-4 pl-8">
            {def.controller ? (
              <Text
                variant="muted"
                as="p"
                className="flex items-center gap-1.5"
              >
                <Server className="h-3.5 w-3.5" />
                Runs on the controller. It takes no target.
              </Text>
            ) : (
              <div className="max-w-xs">
                <TargetPicker value={target} onChange={onTargetChange} />
                {target === "" && (
                  <Text variant="warn" as="p" className="mt-1">
                    Pick the hosts that it changes. It does not run until it has
                    a target.
                  </Text>
                )}
              </div>
            )}
            {children}
            {error && (
              <Text variant="error" as="p">
                {error}
              </Text>
            )}
          </div>
        </Card>
      </div>
    );
  },
);
StepCard.displayName = "StepCard";
