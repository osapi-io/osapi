import { Picker } from "@/components/ui/popover";
import { useObjects } from "@/hooks/use-objects";
import { formatBytes } from "@/lib/format";
import { Database, Upload } from "lucide-react";

interface ObjectPickerProps {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  /** Object names from upstream upload blocks in the stack */
  upstreamObjects?: string[];
}

/** ObjectPicker picks an object of the file store, or one an upload block above makes. */
export function ObjectPicker({
  id,
  label,
  value,
  onChange,
  upstreamObjects = [],
}: ObjectPickerProps) {
  const { objects, loading } = useObjects();

  // The names of upload blocks above, with the objects the server has.
  const names = [
    ...new Set([...upstreamObjects, ...objects.map((o) => o.name)]),
  ].sort();
  const selected = objects.find((o) => o.name === value);

  return (
    <Picker
      id={id}
      label={label}
      value={value}
      onChange={onChange}
      loading={loading}
      placeholder="Select an object"
      options={names.map((name) => {
        const obj = objects.find((o) => o.name === name);
        return {
          value: name,
          icon: upstreamObjects.includes(name) ? (
            <Upload className="h-3 w-3 shrink-0 text-accent-light" />
          ) : (
            <Database className="h-3 w-3 shrink-0 text-text-muted" />
          ),
          meta: obj ? formatBytes(obj.size) : "from upload",
        };
      })}
    >
      {selected && (
        <p className="text-[10px] text-text-muted">
          {formatBytes(selected.size)} &middot; {selected.content_type}
          &middot; sha:{selected.sha256.slice(0, 8)}
        </p>
      )}
      {upstreamObjects.includes(value) && !selected && (
        <p className="text-[10px] text-accent-light">
          Will use object from upload block above
        </p>
      )}
    </Picker>
  );
}
