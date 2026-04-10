import { forwardRef, type InputHTMLAttributes } from "react";
import { cn } from "@/lib/cn";
import { FormField } from "@/components/ui/form-field";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, id, ...props }, ref) => {
    return (
      <FormField id={id} label={label}>
        <input
          ref={ref}
          id={id}
          autoComplete="off"
          data-1p-ignore
          data-lpignore="true"
          className={cn(
            "flex h-9 w-full rounded-md border border-border bg-background px-3 py-1 text-sm text-text",
            "placeholder:text-text-muted",
            "focus:border-primary/40 focus:outline-none focus:ring-1 focus:ring-primary/20",
            className,
          )}
          {...props}
        />
      </FormField>
    );
  },
);
Input.displayName = "Input";
