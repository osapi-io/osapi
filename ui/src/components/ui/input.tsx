import {
  forwardRef,
  type ElementType,
  type InputHTMLAttributes,
  type TextareaHTMLAttributes,
} from "react";
import { cn } from "@/lib/cn";
import { FormField } from "@/components/ui/form-field";

// The look of every text field: its border, its placeholder and its focus
// ring.
const field = cn(
  "w-full rounded-md border border-border bg-background px-3 text-sm text-text",
  "placeholder:text-text-muted",
  "focus:border-primary/40 focus:outline-none focus:ring-1 focus:ring-primary/20",
);

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  /** An icon inside the field, before the text. */
  icon?: ElementType;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, id, icon: Icon, ...props }, ref) => {
    return (
      <FormField id={id} label={label}>
        <div className="relative">
          {Icon && (
            <Icon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-text-muted" />
          )}
          <input
            ref={ref}
            id={id}
            autoComplete="off"
            data-1p-ignore
            data-lpignore="true"
            className={cn("flex h-9 py-1", field, Icon && "pl-10", className)}
            {...props}
          />
        </div>
      </FormField>
    );
  },
);
Input.displayName = "Input";

interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
}

/** Textarea is a text field of several lines, with the look of Input. */
export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, label, id, ...props }, ref) => {
    return (
      <FormField id={id} label={label}>
        <textarea
          ref={ref}
          id={id}
          className={cn("py-2", field, className)}
          {...props}
        />
      </FormField>
    );
  },
);
Textarea.displayName = "Textarea";
