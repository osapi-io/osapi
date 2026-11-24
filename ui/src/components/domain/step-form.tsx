import { Textarea } from "@/components/ui/input";
import { FactInput } from "@/components/ui/fact-input";
import { SecretInput } from "@/components/ui/secret-input";
import { useSecretRefs } from "@/hooks/use-secret-refs";
import { Checkbox } from "@/components/ui/checkbox";
import { Dropdown } from "@/components/ui/dropdown";
import { TagInput } from "@/components/ui/tag-input";
import { Text } from "@/components/ui/text";
import type { CatalogField, CatalogMethod } from "@/lib/catalog";

interface StepFormProps {
  /** The step's id, which keeps its field ids apart from another step's. */
  stepId: string;
  method: CatalogMethod;
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
}

/** The label of a field: its name in words, marked when optional. */
function fieldLabel(f: CatalogField): string {
  const words = f.name.replace(/_/g, " ");
  const label = words.charAt(0).toUpperCase() + words.slice(1);
  return f.required ? label : `${label} (optional)`;
}

/** A list field's value as tags: a list, or text that a person typed. */
function tags(raw: unknown): string[] {
  if (Array.isArray(raw)) return raw.map(String);
  if (typeof raw === "string" && raw.trim() !== "") {
    return raw
      .split(",")
      .map((t) => t.trim())
      .filter(Boolean);
  }
  return [];
}

/**
 * The form of a step, from the input of its method: a select for a field
 * with fixed values, tags for a list, a checkbox for a flag, JSON for an
 * object, a secret field that takes a secret:// reference from a picker,
 * and a text field that takes a @fact. reference for the rest. The
 * doc of each field stays under it, so it does not go when a person types.
 */
export function StepForm({ stepId, method, data, onChange }: StepFormProps) {
  const secrets = useSecretRefs();
  if (method.fields.length === 0) {
    return (
      <Text variant="muted" size="sm" as="p">
        It takes no input.
      </Text>
    );
  }

  const update = (name: string, value: unknown) =>
    onChange({ ...data, [name]: value });

  return (
    <div className="grid gap-x-4 gap-y-3 sm:grid-cols-2">
      {method.fields.map((f) => {
        const id = `${stepId}-${f.name}`;
        const label = fieldLabel(f);
        const value = data[f.name];
        let control;
        if (f.kind === "bool") {
          control = (
            <Checkbox
              id={id}
              label={label}
              checked={value === true}
              onChange={(e) => update(f.name, e.target.checked)}
            />
          );
        } else if (f.options) {
          control = (
            <Dropdown
              id={id}
              label={label}
              value={(value as string) ?? ""}
              placeholder="Pick one"
              onChange={(v) => update(f.name, v)}
              options={[
                ...(f.required ? [] : [{ value: "", label: "Not set" }]),
                ...f.options.map((o) => ({ value: o, label: o })),
              ]}
            />
          );
        } else if (f.kind === "list") {
          control = (
            <TagInput
              id={id}
              label={label}
              value={tags(value)}
              onChange={(v) => update(f.name, v)}
              placeholder="Type a value, then Enter"
            />
          );
        } else if (f.kind === "object") {
          control = (
            <Textarea
              id={id}
              label={label}
              rows={3}
              className="font-mono"
              placeholder='{"key": "value"}'
              value={(value as string) ?? ""}
              onChange={(e) => update(f.name, e.target.value)}
            />
          );
        } else if (f.secret) {
          control = (
            <SecretInput
              id={id}
              label={label}
              value={(value as string) ?? ""}
              onChange={(v) => update(f.name, v)}
              refs={secrets.refs}
              picker={secrets.allowed}
            />
          );
        } else {
          control = (
            <FactInput
              id={id}
              label={label}
              placeholder={f.kind === "number" ? "A number" : undefined}
              value={(value as string) ?? ""}
              onChange={(v) => update(f.name, v)}
            />
          );
        }
        // A list or a JSON object needs the width; a flag is a short line.
        const wide = f.kind === "list" || f.kind === "object";
        return (
          <div key={f.name} className={wide ? "sm:col-span-2" : undefined}>
            {control}
            {f.doc && (
              <Text variant="muted" as="p" className="mt-1">
                {f.doc}
              </Text>
            )}
          </div>
        );
      })}
    </div>
  );
}
