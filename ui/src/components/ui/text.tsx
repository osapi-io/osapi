import { cn } from "@/lib/cn";
import type { ElementType, HTMLAttributes } from "react";

type TextVariant =
  | "default"
  | "muted"
  | "label"
  | "mono"
  | "mono-muted"
  | "mono-primary"
  | "error"
  | "primary"
  | "accent";

type TextSize = "xs" | "sm" | "base";

interface TextProps extends HTMLAttributes<HTMLElement> {
  variant?: TextVariant;
  size?: TextSize;
  as?: ElementType;
  truncate?: boolean;
}

const variantClasses: Record<TextVariant, string> = {
  default: "text-text",
  muted: "text-text-muted",
  label: "font-medium text-text-muted",
  mono: "font-mono text-text",
  "mono-muted": "font-mono text-text-muted",
  "mono-primary": "font-mono text-primary",
  error: "text-status-error",
  primary: "text-primary",
  accent: "text-accent-light",
};

const sizeClasses: Record<TextSize, string> = {
  xs: "text-xs",
  sm: "text-sm",
  base: "text-base",
};

export function Text({
  variant = "default",
  size = "xs",
  as: Component = "span",
  truncate,
  className,
  ...props
}: TextProps) {
  return (
    <Component
      className={cn(
        sizeClasses[size],
        variantClasses[variant],
        truncate && "truncate",
        className,
      )}
      {...props}
    />
  );
}
