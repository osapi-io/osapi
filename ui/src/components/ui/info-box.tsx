import { cn } from "@/lib/cn";
import type { ReactNode } from "react";

interface InfoBoxProps {
  children: ReactNode;
  className?: string;
}

export function InfoBox({ children, className }: InfoBoxProps) {
  return (
    <div className={cn("rounded-md bg-white/[0.02] px-3 py-2", className)}>
      {children}
    </div>
  );
}
