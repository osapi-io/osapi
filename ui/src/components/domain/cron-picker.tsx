import { useEffect, useState, useRef } from "react";
import { Text } from "@/components/ui/text";
import { FormField } from "@/components/ui/form-field";
import { PopoverPanel, PopoverItem } from "@/components/ui/popover";
import { useOutsideClick } from "@/hooks/use-outside-click";
import { usePopoverKeyboard } from "@/hooks/use-popover-keyboard";
import { getNodeScheduleCron } from "@/sdk/gen/schedule-management-api-cron-operations/schedule-management-api-cron-operations";
import type { CronEntry, CronCollectionResponse } from "@/sdk/gen/schemas";
import { ChevronDown, Calendar } from "lucide-react";
import { cn } from "@/lib/cn";

interface CronPickerProps {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
}

export function CronPicker({ id, label, value, onChange }: CronPickerProps) {
  const [entries, setEntries] = useState<CronEntry[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const ref = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  useOutsideClick(ref, () => setOpen(false), open);

  useEffect(() => {
    let mounted = true;
    const fetch = async () => {
      try {
        const result = await getNodeScheduleCron("_all");
        if (mounted && result.status === 200) {
          setEntries((result.data as CronCollectionResponse).results ?? []);
        }
      } catch {
        // silent
      }
      if (mounted) setLoading(false);
    };
    fetch();
    return () => {
      mounted = false;
    };
  }, []);

  // Dedupe cron names across hosts
  const names = [
    ...new Set(entries.map((e) => e.name).filter(Boolean)),
  ] as string[];

  const { activeIndex, handleKeyDown } = usePopoverKeyboard({
    itemCount: names.length,
    onSelect: (i) => {
      onChange(names[i]);
      setOpen(false);
    },
    onClose: () => setOpen(false),
    onOpen: () => setOpen(true),
    open,
    listRef,
  });

  return (
    <FormField id={id} label={label}>
      <div className="relative" ref={ref}>
        <button
          id={id}
          type="button"
          onClick={() => setOpen(!open)}
          onKeyDown={handleKeyDown}
          className={cn(
            "flex h-9 w-full items-center justify-between rounded-md border border-border bg-background px-3 text-sm text-text",
            "focus:border-primary/40 focus:outline-none focus:ring-1 focus:ring-primary/20",
          )}
        >
          <span className={value ? "text-text" : "text-text-muted"}>
            {value || (loading ? "Loading..." : "Select a cron entry")}
          </span>
          <ChevronDown className="h-3.5 w-3.5 text-text-muted" />
        </button>

        {open && (
          <PopoverPanel onClose={() => setOpen(false)}>
            <div ref={listRef}>
              {names.length === 0 && (
                <Text variant="muted" as="p" className="px-3 py-2">
                  No cron entries found
                </Text>
              )}
              {names.map((name, i) => (
                <PopoverItem
                  key={name}
                  selected={value === name}
                  active={i === activeIndex}
                  onClick={() => {
                    onChange(name);
                    setOpen(false);
                  }}
                >
                  <Calendar className="h-3 w-3 shrink-0 text-text-muted" />
                  <span className="text-text">{name}</span>
                </PopoverItem>
              ))}
            </div>
          </PopoverPanel>
        )}
      </div>
    </FormField>
  );
}
