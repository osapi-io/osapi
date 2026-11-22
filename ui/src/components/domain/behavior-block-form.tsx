import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import { FactInput } from "@/components/ui/fact-input";
import { Checkbox } from "@/components/ui/checkbox";
import { Text } from "@/components/ui/text";
import type { BlockStatus } from "@/hooks/use-stack";
import { blockReady, type BehaviorBlockDef } from "@/lib/behavior-block";

interface BehaviorBlockFormProps {
  /** The block's id, which keeps its field ids apart from another block's. */
  blockId: string;
  def: BehaviorBlockDef;
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

/** The label of a field: its name in words, marked when optional. */
function fieldLabel(name: string, required: boolean): string {
  const words = name.replace(/_/g, " ");
  const label = words.charAt(0).toUpperCase() + words.slice(1);
  return required ? label : `${label} (optional)`;
}

/**
 * The form of a behavior block, from the input of its method: a text field
 * for each string, number or list, with its doc as the hint, and a checkbox
 * for each flag. The block is ready when each required field has a value.
 */
export function BehaviorBlockForm({
  blockId,
  def,
  data,
  onChange,
  onStatusChange,
}: BehaviorBlockFormProps) {
  const prevStatus = useRef<BlockStatus | null>(null);
  const ready = blockReady(def, data);

  useEffect(() => {
    const next: BlockStatus = ready ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [ready, onStatusChange]);

  if (def.fields.length === 0) {
    return (
      <Text variant="muted" size="sm">
        It takes no input.
      </Text>
    );
  }

  const update = (name: string, value: unknown) =>
    onChange({ ...data, [name]: value });

  return (
    <div className="space-y-3">
      {def.fields.map((f) => {
        const id = `${blockId}-${f.name}`;
        const label = fieldLabel(f.name, f.required);
        if (f.kind === "bool") {
          return (
            <Checkbox
              key={f.name}
              id={id}
              label={`${label}: ${f.doc}`}
              checked={data[f.name] === true}
              onChange={(e) => update(f.name, e.target.checked)}
            />
          );
        }
        const hint = f.options
          ? `${f.doc} (${f.options.join(", ")})`
          : f.kind === "list"
            ? `${f.doc}, comma separated`
            : f.doc;
        if (f.secret) {
          return (
            <Input
              key={f.name}
              id={id}
              type="password"
              label={label}
              placeholder={hint}
              value={(data[f.name] as string) || ""}
              onChange={(e) => update(f.name, e.target.value)}
            />
          );
        }
        return (
          <FactInput
            key={f.name}
            id={id}
            label={label}
            placeholder={hint}
            value={(data[f.name] as string) || ""}
            onChange={(v) => update(f.name, v)}
          />
        );
      })}
    </div>
  );
}
