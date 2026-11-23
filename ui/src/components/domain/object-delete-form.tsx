import { ObjectPicker } from "@/components/domain/object-picker";

interface ObjectDeleteFormProps {
  stepId: string;
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
}

/** ObjectDeleteForm picks the file of the object store to delete. */
export function ObjectDeleteForm({
  stepId,
  data,
  onChange,
}: ObjectDeleteFormProps) {
  return (
    <ObjectPicker
      id={`${stepId}-name`}
      label="File to delete"
      value={(data.name as string) || ""}
      onChange={(v) => onChange({ ...data, name: v })}
    />
  );
}
