import { cn } from "@/lib/cn";
import { ChevronLeft, ChevronRight } from "lucide-react";
import type { ButtonHTMLAttributes } from "react";

interface ScrollButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  direction: "left" | "right";
}

export function ScrollButton({
  direction,
  className,
  ...props
}: ScrollButtonProps) {
  const Icon = direction === "left" ? ChevronLeft : ChevronRight;

  return (
    <button
      className={cn(
        "flex h-8 w-8 shrink-0 items-center justify-center rounded-md border border-border/60 bg-background text-text-muted transition-colors hover:border-primary/30 hover:text-primary disabled:invisible",
        className,
      )}
      {...props}
    >
      <Icon className="h-4 w-4" />
    </button>
  );
}
