import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";
import type { HTMLAttributes } from "react";

const cardVariants = cva(
  "rounded-lg border bg-card p-4 transition-all duration-200",
  {
    variants: {
      variant: {
        default: [
          "border-border",
          "shadow-[0_1px_2px_rgba(0,0,0,0.4),inset_0_1px_0_rgba(255,255,255,0.03)]",
          "hover:border-primary/40",
          "hover:shadow-[0_4px_16px_rgba(103,234,148,0.12),0_1px_2px_rgba(0,0,0,0.4)]",
        ],
        active: "border-primary/40 shadow-[0_4px_16px_rgba(103,234,148,0.12)]",
        pending: "border-accent/40 shadow-[0_4px_16px_rgba(124,58,237,0.12)]",
        applied:
          "border-status-applied/60 shadow-[0_4px_16px_rgba(34,170,98,0.12)]",
        error:
          "border-status-error/60 shadow-[0_4px_16px_rgba(239,68,68,0.12)]",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);

interface CardProps
  extends HTMLAttributes<HTMLDivElement>, VariantProps<typeof cardVariants> {}

export function Card({ className, variant, ...props }: CardProps) {
  return (
    <div className={cn(cardVariants({ variant }), className)} {...props} />
  );
}

export function CardHeader({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn("flex items-center gap-3 pb-3", className)} {...props} />
  );
}

export function CardTitle({
  className,
  ...props
}: HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h3
      className={cn("text-sm font-semibold text-white", className)}
      {...props}
    />
  );
}

export function CardContent({
  className,
  ...props
}: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("text-sm text-text", className)} {...props} />;
}
