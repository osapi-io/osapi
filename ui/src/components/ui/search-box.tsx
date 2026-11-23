import { useRef, useEffect } from "react";
import { cn } from "@/lib/cn";
import { Search, X } from "lucide-react";
import { IconButton } from "@/components/ui/icon-button";

interface SearchBoxProps {
  value: string;
  onChange: (value: string) => void;
  onClose: () => void;
  placeholder?: string;
  autoFocus?: boolean;
  className?: string;
  /** The width of the field, which is narrow unless a page widens it. */
  inputClassName?: string;
}

export function SearchBox({
  value,
  onChange,
  onClose,
  placeholder = "Search...",
  autoFocus = true,
  className,
  inputClassName = "w-32",
}: SearchBoxProps) {
  const ref = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (autoFocus) ref.current?.focus();
  }, [autoFocus]);

  return (
    <div
      className={cn(
        "flex items-center gap-1 rounded-md border border-border/60 bg-background px-2 py-1",
        className,
      )}
    >
      <Search className="h-3 w-3 text-text-muted" />
      <input
        ref={ref}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        autoComplete="off"
        data-1p-ignore
        data-lpignore="true"
        className={cn(
          "bg-transparent text-xs text-text outline-none placeholder:text-text-muted/60",
          inputClassName,
        )}
      />
      <IconButton icon={X} variant="ghost" size="sm" onClick={onClose} />
    </div>
  );
}
