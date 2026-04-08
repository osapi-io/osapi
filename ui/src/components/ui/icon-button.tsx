import { cn } from "@/lib/cn";
import type { ButtonHTMLAttributes, ElementType } from "react";

type IconButtonVariant = "ghost" | "danger" | "accent";

interface IconButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  icon: ElementType;
  variant?: IconButtonVariant;
  size?: "sm" | "md";
}

const variantClasses: Record<IconButtonVariant, string> = {
  ghost: "text-text-muted hover:bg-white/5 hover:text-text",
  danger: "text-text-muted hover:text-status-error",
  accent: "text-text-muted hover:text-primary",
};

const sizeClasses: Record<"sm" | "md", string> = {
  sm: "rounded p-1",
  md: "rounded-md p-1.5",
};

export function IconButton({
  icon: Icon,
  variant = "ghost",
  size = "sm",
  className,
  ...props
}: IconButtonProps) {
  const iconSize = size === "sm" ? "h-4 w-4" : "h-4 w-4";

  return (
    <button
      className={cn(
        "transition-colors",
        sizeClasses[size],
        variantClasses[variant],
        className,
      )}
      {...props}
    >
      <Icon className={iconSize} />
    </button>
  );
}
