import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";
import type { HTMLAttributes } from "react";

const badgeVariants = cva(
  "inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium",
  {
    variants: {
      variant: {
        ready: "bg-primary/10 text-primary border border-primary/20",
        pending: "bg-accent/10 text-accent-light border border-accent/20",
        running:
          "bg-status-running/10 text-status-running border border-status-running/20",
        error:
          "bg-status-error/10 text-status-error border border-status-error/20",
        applied:
          "bg-status-applied/10 text-status-applied border border-status-applied/20",
        muted: "bg-text-muted/10 text-text-muted border border-text-muted/20",
      },
    },
    defaultVariants: {
      variant: "muted",
    },
  },
);

interface BadgeProps
  extends HTMLAttributes<HTMLSpanElement>, VariantProps<typeof badgeVariants> {}

export function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <span className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}
