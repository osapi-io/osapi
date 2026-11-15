import { cn } from "@/lib/cn";
import type { ElementType, ReactNode } from "react";

interface SectionLabelProps {
  children: ReactNode;
  icon?: ElementType;
  className?: string;
}

export function SectionLabel({
  children,
  icon: Icon,
  className,
}: SectionLabelProps) {
  return (
    // PageLayout colors each heading by its place on the page.
    <h2
      data-section-heading
      className={cn(
        "mb-2 flex items-center gap-2 text-sm font-semibold text-heading-1",
        className,
      )}
    >
      {/* The logo's dark blue, as an accent: too dark to read as text. */}
      <span aria-hidden className="h-3.5 w-1 rounded-sm bg-accent" />
      {Icon && <Icon className="h-3.5 w-3.5" />}
      {children}
    </h2>
  );
}
