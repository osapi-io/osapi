import { useState, useRef, type ReactNode } from "react";
import { FormField } from "@/components/ui/form-field";
import { Spinner } from "@/components/ui/spinner";
import { Text } from "@/components/ui/text";
import { useOutsideClick } from "@/hooks/use-outside-click";
import { usePopoverKeyboard } from "@/hooks/use-popover-keyboard";
import { cn } from "@/lib/cn";
import { ChevronDown } from "lucide-react";

/** One choice of a Picker. */
export interface PickerOption {
  value: string;
  /** What the choice shows; the value when not given. */
  label?: ReactNode;
  /** An icon before the label. */
  icon?: ReactNode;
  /** Gray text after the label, such as a size or a host. */
  meta?: ReactNode;
  /** The group heading the choice goes under, in the order first seen. */
  group?: string;
  /** A key when two choices share a value, such as one name on two hosts. */
  key?: string;
}

interface PickerProps {
  id?: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  options: PickerOption[];
  /** What the trigger shows when no choice is picked. */
  placeholder: string;
  /** The trigger while the choices load. */
  loading?: boolean;
  /** What the trigger shows for the picked value; the value when not given. */
  display?: ReactNode;
  className?: string;
  panelClassName?: string;
  /** What shows under the trigger, such as the picked choice's details. */
  children?: ReactNode;
}

/**
 * Picker is a labeled field that picks one value from a list: the trigger,
 * the popover with its groups, the keyboard, the loading state and the
 * empty list, which says None. as the CLI does. Each picker of a resource,
 * such as an object or a container, gives only its options.
 */
export function Picker({
  id,
  label,
  value,
  onChange,
  options,
  placeholder,
  loading,
  display,
  className,
  panelClassName,
  children,
}: PickerProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  useOutsideClick(ref, () => setOpen(false), open);

  const pick = (v: string) => {
    onChange(v);
    setOpen(false);
  };
  const { activeIndex, handleKeyDown } = usePopoverKeyboard({
    itemCount: options.length,
    onSelect: (i) => pick(options[i].value),
    onClose: () => setOpen(false),
    onOpen: () => setOpen(true),
    open,
    listRef,
  });

  return (
    <FormField id={id} label={label} className={className}>
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
          <span
            className={cn(
              "flex items-center gap-2",
              value ? "text-text" : "text-text-muted",
            )}
          >
            {value ? (
              (display ?? value)
            ) : loading ? (
              <Spinner size="inline" />
            ) : (
              placeholder
            )}
          </span>
          <ChevronDown className="h-3.5 w-3.5 text-text-muted" />
        </button>

        {open && (
          <PopoverPanel
            onClose={() => setOpen(false)}
            className={panelClassName}
          >
            <div ref={listRef}>
              {options.length === 0 && (
                <Text variant="muted" as="p" className="px-3 py-2">
                  None.
                </Text>
              )}
              {options.map((o, i) => (
                <div key={o.key ?? o.value}>
                  {o.group && o.group !== options[i - 1]?.group && (
                    <p className="px-3 pb-1 pt-2 text-xs font-semibold text-text-muted">
                      {o.group}
                    </p>
                  )}
                  <PopoverItem
                    selected={value === o.value}
                    active={i === activeIndex}
                    onClick={() => pick(o.value)}
                  >
                    {o.icon}
                    <span className="flex-1 truncate text-text">
                      {o.label ?? o.value}
                    </span>
                    {o.meta && (
                      <span className="text-[10px] text-text-muted">
                        {o.meta}
                      </span>
                    )}
                  </PopoverItem>
                </div>
              ))}
            </div>
          </PopoverPanel>
        )}
      </div>
      {children}
    </FormField>
  );
}

interface PopoverPanelProps {
  children: ReactNode;
  onClose: () => void;
  className?: string;
}

// A panel opens above the sticky bars of a page, such as the Run bar of
// Plans (z-20), and under a modal or the command bar (z-50). At z-20 the
// Run bar, which comes later on the page, hid its middle rows (#1268).
export function PopoverPanel({ children, className }: PopoverPanelProps) {
  return (
    <div
      className={cn(
        "absolute z-30 mt-1 max-h-48 w-full overflow-auto rounded-md border border-border bg-card shadow-lg",
        className,
      )}
    >
      {children}
    </div>
  );
}

interface PopoverItemProps {
  selected?: boolean;
  active?: boolean;
  onClick: () => void;
  children: ReactNode;
  className?: string;
}

export function PopoverItem({
  selected,
  active,
  onClick,
  children,
  className,
}: PopoverItemProps) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "flex w-full items-center gap-2 px-3 py-2 text-left text-xs transition-colors hover:bg-primary/5",
        active
          ? "bg-primary/20 border-l-2 border-primary"
          : "border-l-2 border-transparent",
        selected && !active && "bg-primary/10",
        className,
      )}
    >
      {children}
    </button>
  );
}
