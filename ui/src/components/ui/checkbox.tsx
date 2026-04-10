import { forwardRef, type InputHTMLAttributes } from "react";
import { cn } from "@/lib/cn";
import { Check } from "lucide-react";

interface CheckboxProps extends Omit<
  InputHTMLAttributes<HTMLInputElement>,
  "type"
> {
  label?: string;
}

export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
  ({ className, label, id, checked, ...props }, ref) => {
    return (
      <label htmlFor={id} className="flex cursor-pointer items-center gap-2.5">
        <div className="relative">
          <input
            ref={ref}
            type="checkbox"
            id={id}
            checked={checked}
            className="peer sr-only"
            {...props}
          />
          <div
            className={cn(
              "flex h-5 w-5 items-center justify-center rounded border transition-colors",
              checked
                ? "border-primary bg-primary"
                : "border-accent bg-transparent",
              className,
            )}
          >
            {checked && <Check className="h-3.5 w-3.5 text-black" />}
          </div>
        </div>
        {label && (
          <span className="text-sm font-medium text-text">{label}</span>
        )}
      </label>
    );
  },
);
Checkbox.displayName = "Checkbox";
