import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import { FactInput } from "@/components/ui/fact-input";
import type { BlockStatus } from "@/hooks/use-stack";

interface SingleInputBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
  field: string;
  label: string;
  placeholder: string;
  /** Enable @fact. reference suggestions */
  facts?: boolean;
}

export function SingleInputBlock({
  data,
  onChange,
  onStatusChange,
  field,
  label,
  placeholder,
  facts,
}: SingleInputBlockProps) {
  const value = (data[field] as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus = value.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [value, onStatusChange]);

  if (facts) {
    return (
      <FactInput
        id={`${field}-input`}
        label={label}
        placeholder={placeholder}
        value={value}
        onChange={(v) => onChange({ ...data, [field]: v })}
      />
    );
  }

  return (
    <Input
      id={`${field}-input`}
      label={label}
      placeholder={placeholder}
      value={value}
      onChange={(e) => onChange({ ...data, [field]: e.target.value })}
    />
  );
}
