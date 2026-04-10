import { cn } from "@/lib/cn";
import { ChevronDown, ChevronRight } from "lucide-react";
import type { ElementType, ReactNode } from "react";

interface CollapsibleSectionProps {
  open: boolean;
  onToggle: () => void;
  label: ReactNode;
  icon?: ElementType;
  rightContent?: ReactNode;
  children: ReactNode;
  className?: string;
}

export function CollapsibleSection({
  open,
  onToggle,
  label,
  icon: Icon,
  rightContent,
  children,
  className,
}: CollapsibleSectionProps) {
  return (
    <div className={cn("border-t border-border/30", className)}>
      <button
        onClick={onToggle}
        className="flex w-full items-center gap-1.5 px-4 py-1.5 text-left text-xs text-text-muted hover:text-text"
      >
        {Icon && <Icon className="h-3 w-3" />}
        {label}
        {rightContent && <span className="ml-2">{rightContent}</span>}
        {open ? (
          <ChevronDown className="ml-auto h-3 w-3" />
        ) : (
          <ChevronRight className="ml-auto h-3 w-3" />
        )}
      </button>
      {open && children}
    </div>
  );
}
