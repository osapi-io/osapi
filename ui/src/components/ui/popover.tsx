import { useState, useRef, useEffect, type ReactNode } from "react";
import { cn } from "@/lib/cn";
import { ChevronDown } from "lucide-react";

interface PopoverProps {
  /** Currently selected display text */
  value?: ReactNode;
  /** Placeholder when no value selected */
  placeholder?: string;
  /** Whether to highlight the trigger (e.g. for @fact. values) */
  highlight?: boolean;
  /** Popover panel content */
  children: ReactNode;
  /** Additional trigger classes */
  className?: string;
}

export function Popover({
  value,
  placeholder = "Select...",
  highlight,
  children,
  className,
}: PopoverProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className={cn(
          "flex h-9 w-full items-center justify-between rounded-md border border-border bg-background px-3 text-sm text-text",
          "focus:border-primary/40 focus:outline-none focus:ring-1 focus:ring-primary/20",
          className,
        )}
      >
        <span
          className={
            value
              ? highlight
                ? "text-accent-light"
                : "text-text"
              : "text-text-muted"
          }
        >
          {value ?? placeholder}
        </span>
        <ChevronDown className="h-3.5 w-3.5 text-text-muted" />
      </button>
      {open && (
        <PopoverPanel onClose={() => setOpen(false)}>{children}</PopoverPanel>
      )}
    </div>
  );
}

interface PopoverPanelProps {
  children: ReactNode;
  onClose: () => void;
  className?: string;
}

export function PopoverPanel({ children, className }: PopoverPanelProps) {
  return (
    <div
      className={cn(
        "absolute z-20 mt-1 max-h-48 w-full overflow-auto rounded-md border border-border bg-card shadow-lg",
        className,
      )}
    >
      {children}
    </div>
  );
}

interface PopoverItemProps {
  selected?: boolean;
  active?: boolean;
  onClick: () => void;
  children: ReactNode;
  className?: string;
}

export function PopoverItem({
  selected,
  active,
  onClick,
  children,
  className,
}: PopoverItemProps) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "flex w-full items-center gap-2 px-3 py-2 text-left text-xs transition-colors hover:bg-primary/5",
        active
          ? "bg-primary/20 border-l-2 border-primary"
          : "border-l-2 border-transparent",
        selected && !active && "bg-primary/10",
        className,
      )}
    >
      {children}
    </button>
  );
}
