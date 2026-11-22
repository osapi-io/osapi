import { cn } from "@/lib/cn";
import type { ButtonHTMLAttributes } from "react";

interface ChipProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  /** The chip is the one picked, such as the open group. */
  active?: boolean;
  size?: "sm" | "md";
}

/**
 * Chip is a bordered choice in a row of them, such as a block group or a
 * block type on the Configure page. A disabled chip is faint, for a choice
 * the token's role cannot use.
 */
export function Chip({
  active,
  size = "md",
  disabled,
  className,
  ...props
}: ChipProps) {
  return (
    <button
      type="button"
      disabled={disabled}
      className={cn(
        "flex items-center gap-1.5 rounded-md border font-medium transition-colors",
        size === "sm" ? "px-3 py-1.5 text-xs" : "px-3 py-2 text-sm",
        disabled
          ? "cursor-not-allowed border-border/30 bg-card/50 text-text-muted/40"
          : active
            ? // The one picked: the border and fill that a chip under the
              // pointer has, a little stronger, and the white text. No
              // colored line: the docs mark the open page the same way.
              // Green is a status, not a selection.
              "border-white/20 bg-white/[0.06] text-white"
            : "border-border bg-card text-text hover:border-white/20 hover:bg-white/[0.03]",
        className,
      )}
      {...props}
    />
  );
}
