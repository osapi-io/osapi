import { useEffect, useState, useRef } from "react";
import { Text } from "@/components/ui/text";
import { FormField } from "@/components/ui/form-field";
import { PopoverPanel, PopoverItem } from "@/components/ui/popover";
import { useOutsideClick } from "@/hooks/use-outside-click";
import { usePopoverKeyboard } from "@/hooks/use-popover-keyboard";
import { getNodeContainerDocker } from "@/sdk/gen/docker-management-api-docker-operations/docker-management-api-docker-operations";
import type {
  DockerListCollectionResponse,
  DockerSummary,
} from "@/sdk/gen/schemas";
import { ChevronDown, Container } from "lucide-react";
import { cn } from "@/lib/cn";

interface ContainerPickerProps {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
}

interface FlatContainer extends DockerSummary {
  hostname: string;
}

export function ContainerPicker({
  id,
  label,
  value,
  onChange,
}: ContainerPickerProps) {
  const [containers, setContainers] = useState<FlatContainer[]>([]);
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const ref = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  useOutsideClick(ref, () => setOpen(false), open);

  useEffect(() => {
    let mounted = true;
    const fetch = async () => {
      try {
        const result = await getNodeContainerDocker("_all");
        if (mounted && result.status === 200) {
          const data = result.data as DockerListCollectionResponse;
          const flat: FlatContainer[] = [];
          for (const host of data.results) {
            for (const c of host.containers ?? []) {
              flat.push({ ...c, hostname: host.hostname });
            }
          }
          setContainers(flat);
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

  const { activeIndex, handleKeyDown } = usePopoverKeyboard({
    itemCount: containers.length,
    onSelect: (i) => {
      onChange(containers[i].name || containers[i].id || "");
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
            {value || (loading ? "Loading..." : "Select a container")}
          </span>
          <ChevronDown className="h-3.5 w-3.5 text-text-muted" />
        </button>

        {open && (
          <PopoverPanel onClose={() => setOpen(false)} className="max-h-60">
            <div ref={listRef}>
              {containers.length === 0 && (
                <Text variant="muted" as="p" className="px-3 py-2">
                  No containers found
                </Text>
              )}
              {containers.map((c, i) => {
                const id_or_name = c.name || c.id || "";
                return (
                  <PopoverItem
                    key={`${c.hostname}-${c.id}`}
                    selected={value === id_or_name}
                    active={i === activeIndex}
                    onClick={() => {
                      onChange(id_or_name);
                      setOpen(false);
                    }}
                  >
                    <Container className="h-3 w-3 shrink-0 text-text-muted" />
                    <span className="flex-1 truncate text-text">
                      {c.name || c.id?.slice(0, 12)}
                    </span>
                    <span className="text-[10px] text-text-muted">
                      {c.hostname}
                    </span>
                    {c.state && (
                      <span
                        className={`text-[10px] ${c.state === "running" ? "text-primary" : "text-text-muted"}`}
                      >
                        {c.state}
                      </span>
                    )}
                  </PopoverItem>
                );
              })}
            </div>
          </PopoverPanel>
        )}
      </div>
    </FormField>
  );
}
