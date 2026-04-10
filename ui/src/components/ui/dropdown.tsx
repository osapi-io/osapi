import { useState, useRef } from "react";
import { FormField } from "@/components/ui/form-field";
import { PopoverPanel, PopoverItem } from "@/components/ui/popover";
import { useOutsideClick } from "@/hooks/use-outside-click";
import { ChevronDown } from "lucide-react";
import { cn } from "@/lib/cn";

export interface DropdownOption {
  value: string;
  label: string;
}

interface DropdownProps {
  id?: string;
  label?: string;
  value: string;
  options: DropdownOption[];
  placeholder?: string;
  onChange: (value: string) => void;
  className?: string;
}

export function Dropdown({
  id,
  label,
  value,
  options,
  placeholder = "Select...",
  onChange,
  className,
}: DropdownProps) {
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(-1);
  const ref = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  useOutsideClick(
    ref,
    () => {
      setOpen(false);
      setActiveIndex(-1);
    },
    open,
  );

  const selected = options.find((o) => o.value === value);

  const select = (opt: DropdownOption) => {
    onChange(opt.value);
    setOpen(false);
    setActiveIndex(-1);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!open) {
      if (e.key === "ArrowDown" || e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        setOpen(true);
        const idx = options.findIndex((o) => o.value === value);
        setActiveIndex(idx >= 0 ? idx : 0);
      }
      return;
    }

    if (e.key === "Escape") {
      e.preventDefault();
      setOpen(false);
      setActiveIndex(-1);
    } else if (e.key === "ArrowDown") {
      e.preventDefault();
      const next = activeIndex < options.length - 1 ? activeIndex + 1 : 0;
      setActiveIndex(next);
      const buttons = listRef.current?.querySelectorAll("button");
      buttons?.[next]?.scrollIntoView({ block: "nearest" });
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      const next = activeIndex > 0 ? activeIndex - 1 : options.length - 1;
      setActiveIndex(next);
      const buttons = listRef.current?.querySelectorAll("button");
      buttons?.[next]?.scrollIntoView({ block: "nearest" });
    } else if (e.key === "Enter" && activeIndex >= 0) {
      e.preventDefault();
      select(options[activeIndex]);
    }
  };

  return (
    <FormField id={id} label={label} className={className}>
      <div className="relative" ref={ref}>
        <button
          id={id}
          type="button"
          onClick={() => {
            setOpen(!open);
            if (!open) {
              const idx = options.findIndex((o) => o.value === value);
              setActiveIndex(idx >= 0 ? idx : 0);
            }
          }}
          onKeyDown={handleKeyDown}
          className={cn(
            "flex h-9 w-full items-center justify-between rounded-md border border-border bg-background px-3 text-sm text-text",
            "focus:border-primary/40 focus:outline-none focus:ring-1 focus:ring-primary/20",
          )}
        >
          <span className={selected ? "text-text" : "text-text-muted"}>
            {selected?.label ?? placeholder}
          </span>
          <ChevronDown className="h-3.5 w-3.5 text-text-muted" />
        </button>

        {open && (
          <PopoverPanel onClose={() => setOpen(false)}>
            <div ref={listRef}>
              {options.map((opt, i) => (
                <PopoverItem
                  key={opt.value}
                  selected={value === opt.value}
                  active={i === activeIndex}
                  onClick={() => select(opt)}
                >
                  <span className="text-text">{opt.label}</span>
                </PopoverItem>
              ))}
            </div>
          </PopoverPanel>
        )}
      </div>
    </FormField>
  );
}
