import { useEffect, useState, useRef } from "react";
import { Text } from "@/components/ui/text";
import { FormField } from "@/components/ui/form-field";
import { PopoverPanel, PopoverItem } from "@/components/ui/popover";
import { useOutsideClick } from "@/hooks/use-outside-click";
import { usePopoverKeyboard } from "@/hooks/use-popover-keyboard";
import { getFiles } from "@/sdk/gen/file-management-api-file-operations/file-management-api-file-operations";
import type { FileInfo, FileListResponse } from "@/sdk/gen/schemas";
import { ChevronDown, Database, Upload } from "lucide-react";
import { cn } from "@/lib/cn";

interface ObjectPickerProps {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  /** Object names from upstream upload blocks in the stack */
  upstreamObjects?: string[];
}

export function ObjectPicker({
  id,
  label,
  value,
  onChange,
  upstreamObjects = [],
}: ObjectPickerProps) {
  const [objects, setObjects] = useState<FileInfo[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const ref = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  useOutsideClick(ref, () => setOpen(false), open);

  useEffect(() => {
    let mounted = true;
    const fetch = async () => {
      try {
        const result = await getFiles();
        if (mounted && result.status === 200) {
          setObjects((result.data as FileListResponse).files ?? []);
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

  // Merge upstream names (from upload blocks above) with server objects
  const allNames = new Set<string>();
  for (const name of upstreamObjects) allNames.add(name);
  for (const obj of objects) allNames.add(obj.name);
  const options = Array.from(allNames).sort();

  const selectedObj = objects.find((o) => o.name === value);
  const isUpstream = upstreamObjects.includes(value);

  const { activeIndex, handleKeyDown } = usePopoverKeyboard({
    itemCount: options.length,
    onSelect: (i) => {
      onChange(options[i]);
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
            {value || (loading ? "Loading..." : "Select an object")}
          </span>
          <ChevronDown className="h-3.5 w-3.5 text-text-muted" />
        </button>

        {open && (
          <PopoverPanel onClose={() => setOpen(false)}>
            <div ref={listRef}>
              {options.length === 0 && (
                <Text variant="muted" as="p" className="px-3 py-2">
                  No objects available
                </Text>
              )}
              {options.map((name, i) => {
                const obj = objects.find((o) => o.name === name);
                const fromUpstream = upstreamObjects.includes(name);
                return (
                  <PopoverItem
                    key={name}
                    selected={value === name}
                    active={i === activeIndex}
                    onClick={() => {
                      onChange(name);
                      setOpen(false);
                    }}
                  >
                    {fromUpstream ? (
                      <Upload className="h-3 w-3 shrink-0 text-accent-light" />
                    ) : (
                      <Database className="h-3 w-3 shrink-0 text-text-muted" />
                    )}
                    <span className="flex-1 truncate text-text">{name}</span>
                    {obj && (
                      <span className="text-[10px] text-text-muted">
                        {formatSize(obj.size)}
                      </span>
                    )}
                    {fromUpstream && !obj && (
                      <span className="text-[10px] text-accent-light">
                        from upload
                      </span>
                    )}
                  </PopoverItem>
                );
              })}
            </div>
          </PopoverPanel>
        )}
      </div>

      {/* Show details for selected */}
      {selectedObj && (
        <p className="text-[10px] text-text-muted">
          {formatSize(selectedObj.size)} &middot; {selectedObj.content_type}
          &middot; sha:{selectedObj.sha256.slice(0, 8)}
        </p>
      )}
      {isUpstream && !selectedObj && (
        <p className="text-[10px] text-accent-light">
          Will use object from upload block above
        </p>
      )}
    </FormField>
  );
}

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
