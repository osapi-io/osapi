import { cn } from "@/lib/cn";
import { methodKindLabels, type MethodKind } from "@/lib/catalog";

// The logo's colors, not the status colors: green, amber and red mean a
// status, and a write is not a warning.
const tones: Record<MethodKind, string> = {
  read: "border-text-muted/40 text-text-muted",
  write: "border-accent-light/60 text-accent-light",
  action: "border-heading-2/60 text-heading-2",
};

const letters: Record<MethodKind, string> = {
  read: "R",
  write: "W",
  action: "A",
};

/**
 * KindMark marks a method as a read (R), a write (W) or an action (A): what
 * it does to a host. A locked mark is faint, for a method that the viewer's
 * role cannot run.
 */
export function KindMark({
  kind,
  locked,
  title,
  className,
}: {
  kind: MethodKind;
  locked?: boolean;
  /** The tooltip, such as the method's name; the kind when not given. */
  title?: string;
  className?: string;
}) {
  return (
    <span
      role="img"
      aria-label={title ?? methodKindLabels[kind]}
      title={title ?? methodKindLabels[kind]}
      className={cn(
        "inline-grid h-4 w-4 shrink-0 place-items-center rounded-[3px] border font-mono text-[10px] font-bold leading-none",
        tones[kind],
        locked && "opacity-30",
        className,
      )}
    >
      {letters[kind]}
    </span>
  );
}

/** KindLabel is a mark with its word: R Read. */
export function KindLabel({ kind }: { kind: MethodKind }) {
  return (
    <span className="inline-flex items-center gap-1.5 text-xs text-text-muted">
      <KindMark kind={kind} />
      {methodKindLabels[kind]}
    </span>
  );
}
