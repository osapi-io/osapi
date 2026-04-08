import { cn } from "@/lib/cn";
import type { ReactNode } from "react";

interface FormFieldProps {
  id?: string;
  label?: string;
  children: ReactNode;
  className?: string;
}

export function FormField({ id, label, children, className }: FormFieldProps) {
  return (
    <div className={cn("space-y-1.5", className)}>
      {label && (
        <label htmlFor={id} className="text-xs font-medium text-text-muted">
          {label}
        </label>
      )}
      {children}
    </div>
  );
}
