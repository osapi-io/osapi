import { useState, useRef } from "react";
import { cn } from "@/lib/cn";
import { FormField } from "@/components/ui/form-field";
import { PopoverPanel, PopoverItem } from "@/components/ui/popover";
import { useOutsideClick } from "@/hooks/use-outside-click";
import { useFacts } from "@/hooks/use-facts";
import { AtSign, ChevronDown } from "lucide-react";

interface FactInputProps {
  id?: string;
  label?: string;
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
  className?: string;
}

export function FactInput({
  id,
  label,
  placeholder,
  value,
  onChange,
  className,
}: FactInputProps) {
  const { facts } = useFacts();
  const [open, setOpen] = useState(false);
  const [filter, setFilter] = useState("");
  const [activeIndex, setActiveIndex] = useState(-1);
  const ref = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  const isFact =
    value.startsWith("@fact.") ||
    (value.length > 0 && "@fact.".startsWith(value));

  useOutsideClick(
    ref,
    () => {
      setOpen(false);
      setFilter("");
      setActiveIndex(-1);
    },
    open,
  );

  const filtered = filter
    ? facts.filter(
        (f) =>
          f.key.startsWith(filter.toLowerCase()) ||
          (f.description ?? "").toLowerCase().startsWith(filter.toLowerCase()),
      )
    : facts;

  const handleSelect = (key: string) => {
    onChange(`@fact.${key}`);
    setOpen(false);
    setFilter("");
    setActiveIndex(-1);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const v = e.target.value;
    onChange(v);
    setActiveIndex(-1);

    // Auto-open suggestions while typing toward @fact.
    if ("@fact.".startsWith(v) || v.startsWith("@fact.")) {
      setOpen(true);
      setFilter(v.startsWith("@fact.") ? v.slice(6) : "");
    } else {
      setOpen(false);
      setFilter("");
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!open || filtered.length === 0) return;

    if (e.key === "ArrowDown") {
      e.preventDefault();
      const next = activeIndex < filtered.length - 1 ? activeIndex + 1 : 0;
      setActiveIndex(next);
      listRef.current?.children[next]?.scrollIntoView({ block: "nearest" });
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      const next = activeIndex > 0 ? activeIndex - 1 : filtered.length - 1;
      setActiveIndex(next);
      listRef.current?.children[next]?.scrollIntoView({ block: "nearest" });
    } else if (e.key === "Enter" && activeIndex >= 0) {
      e.preventDefault();
      handleSelect(filtered[activeIndex].key);
    } else if (e.key === "Tab") {
      // Tab-complete only when user has typed a filter and it narrows results,
      // or when they've explicitly arrowed to an item
      if (activeIndex >= 0) {
        e.preventDefault();
        handleSelect(filtered[activeIndex].key);
      } else if (filter && filtered.length < facts.length) {
        e.preventDefault();
        handleSelect(filtered[0].key);
      }
    } else if (e.key === "Escape") {
      setOpen(false);
      setFilter("");
      setActiveIndex(-1);
    }
  };

  return (
    <FormField id={id} label={label} className={className}>
      <div className="relative" ref={ref}>
        <input
          ref={inputRef}
          id={id}
          value={value}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          autoComplete="off"
          data-1p-ignore
          data-lpignore="true"
          className={cn(
            "flex h-9 w-full rounded-md border border-border bg-background px-3 py-1 pr-8 text-sm text-text",
            "placeholder:text-text-muted",
            "focus:border-primary/40 focus:outline-none focus:ring-1 focus:ring-primary/20",
            isFact && "text-accent-light",
          )}
        />
        <button
          type="button"
          onClick={() => {
            setOpen(!open);
            setFilter("");
            setActiveIndex(-1);
            inputRef.current?.focus();
          }}
          className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-0.5 text-text-muted hover:text-primary"
          title="Insert @fact reference"
        >
          {open ? (
            <ChevronDown className="h-3.5 w-3.5" />
          ) : (
            <AtSign className="h-3.5 w-3.5" />
          )}
        </button>

        {open && (
          <PopoverPanel
            onClose={() => {
              setOpen(false);
              setFilter("");
              setActiveIndex(-1);
            }}
          >
            <div className="sticky top-0 border-b border-border/30 bg-card px-3 py-1.5">
              <p className="text-xs font-semibold text-text-muted">
                @fact references
              </p>
            </div>
            <div ref={listRef}>
              {filtered.length === 0 ? (
                <p className="px-3 py-2 text-xs text-text-muted">
                  No matching facts
                </p>
              ) : (
                filtered.map((f, i) => (
                  <PopoverItem
                    key={f.key}
                    selected={value === `@fact.${f.key}`}
                    active={activeIndex === i}
                    onClick={() => handleSelect(f.key)}
                    className="flex-col items-start gap-0"
                  >
                    <span className="text-xs font-medium text-accent-light">
                      @fact.{f.key}
                    </span>
                    <span className="text-xs text-text-muted">
                      {f.description}
                    </span>
                  </PopoverItem>
                ))
              )}
            </div>
          </PopoverPanel>
        )}
      </div>
    </FormField>
  );
}
