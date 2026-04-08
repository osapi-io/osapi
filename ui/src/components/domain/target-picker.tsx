import { useState, useRef, useMemo } from "react";
import { useTargets, type TargetOption } from "@/hooks/use-targets";
import { FormField } from "@/components/ui/form-field";
import { PopoverPanel, PopoverItem } from "@/components/ui/popover";
import { useOutsideClick } from "@/hooks/use-outside-click";
import { usePopoverKeyboard } from "@/hooks/use-popover-keyboard";
import { cn } from "@/lib/cn";
import { ChevronDown, Globe, Monitor, Tag } from "lucide-react";

interface TargetPickerProps {
  value: string;
  onChange: (target: string) => void;
  label?: string;
  className?: string;
}

function GroupIcon({ group }: { group: TargetOption["group"] }) {
  switch (group) {
    case "builtin":
      return <Globe className="h-3 w-3 shrink-0 text-primary" />;
    case "hostname":
      return <Monitor className="h-3 w-3 shrink-0 text-accent-light" />;
    case "label":
      return <Tag className="h-3 w-3 shrink-0 text-status-running" />;
  }
}

function groupLabel(group: TargetOption["group"]) {
  switch (group) {
    case "builtin":
      return "Targets";
    case "hostname":
      return "Hosts";
    case "label":
      return "Labels";
  }
}

export function TargetPicker({
  value,
  onChange,
  label = "Target",
  className,
}: TargetPickerProps) {
  const { options } = useTargets();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  useOutsideClick(ref, () => setOpen(false), open);

  const grouped = options.reduce<Record<string, TargetOption[]>>((acc, opt) => {
    (acc[opt.group] ??= []).push(opt);
    return acc;
  }, {});

  const flatOptions = useMemo(() => Object.values(grouped).flat(), [grouped]);

  const selected = options.find((o) => o.value === value);

  const { activeIndex, handleKeyDown } = usePopoverKeyboard({
    itemCount: flatOptions.length,
    onSelect: (i) => {
      onChange(flatOptions[i].value);
      setOpen(false);
    },
    onClose: () => setOpen(false),
    onOpen: () => setOpen(true),
    open,
    listRef,
  });

  let flatIdx = 0;

  return (
    <FormField label={label} className={className}>
      <div className="relative" ref={ref}>
        <button
          type="button"
          onClick={() => setOpen(!open)}
          onKeyDown={handleKeyDown}
          className={cn(
            "flex h-9 w-full items-center justify-between rounded-md border border-border bg-background px-3 text-sm text-text",
            "focus:border-primary/40 focus:outline-none focus:ring-1 focus:ring-primary/20",
          )}
        >
          <span className="flex items-center gap-2">
            {selected && <GroupIcon group={selected.group} />}
            <span className={value ? "text-text" : "text-text-muted"}>
              {selected?.label ?? value}
            </span>
          </span>
          <ChevronDown className="h-3.5 w-3.5 text-text-muted" />
        </button>

        {open && (
          <PopoverPanel onClose={() => setOpen(false)} className="max-h-56">
            <div ref={listRef}>
              {Object.entries(grouped).map(([group, opts]) => (
                <div key={group}>
                  <p className="px-3 pb-1 pt-2 text-xs font-semibold uppercase tracking-wider text-text-muted">
                    {groupLabel(group as TargetOption["group"])}
                  </p>
                  {opts.map((opt) => {
                    const idx = flatIdx++;
                    return (
                      <PopoverItem
                        key={opt.value}
                        selected={value === opt.value}
                        active={idx === activeIndex}
                        onClick={() => {
                          onChange(opt.value);
                          setOpen(false);
                        }}
                      >
                        <GroupIcon group={opt.group} />
                        <span className="flex-1 truncate text-text">
                          {opt.label}
                        </span>
                      </PopoverItem>
                    );
                  })}
                </div>
              ))}
            </div>
          </PopoverPanel>
        )}
      </div>
    </FormField>
  );
}
