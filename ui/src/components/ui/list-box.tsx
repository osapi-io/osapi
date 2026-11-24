import type { HTMLAttributes } from "react";
import { cn } from "@/lib/cn";

interface ListBoxProps extends HTMLAttributes<HTMLUListElement> {
  /** A rule between two rows. */
  divided?: boolean;
}

/** A list of rows in one box, as the behaviors and the step picker show. */
export function ListBox({ divided, className, ...props }: ListBoxProps) {
  return (
    <ul
      className={cn(
        "overflow-hidden rounded-lg border border-border bg-card",
        divided && "divide-y divide-border",
        className,
      )}
      {...props}
    />
  );
}
