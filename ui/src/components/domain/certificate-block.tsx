import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface CertificateBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
  /** When true, shows update fields (object only, no name input) */
  mode?: "create" | "update";
}

export function CertificateBlock({
  data,
  onChange,
  onStatusChange,
  mode = "create",
}: CertificateBlockProps) {
  const name = (data.name as string) || "";
  const object = (data.object as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const ready =
      mode === "update"
        ? object.trim() !== ""
        : name.trim() !== "" && object.trim() !== "";
    const next: BlockStatus = ready ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [name, object, mode, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  if (mode === "update") {
    return (
      <div className="grid grid-cols-2 gap-3">
        <Input
          id="cert-name"
          label="Certificate Name"
          placeholder="my-ca"
          value={name}
          onChange={(e) => update("name", e.target.value)}
        />
        <Input
          id="cert-object"
          label="Object Store Reference"
          placeholder="ca-cert.pem"
          value={object}
          onChange={(e) => update("object", e.target.value)}
        />
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 gap-3">
      <Input
        id="cert-name"
        label="Certificate Name"
        placeholder="my-ca"
        value={name}
        onChange={(e) => update("name", e.target.value)}
      />
      <Input
        id="cert-object"
        label="Object Store Reference"
        placeholder="ca-cert.pem"
        value={object}
        onChange={(e) => update("object", e.target.value)}
      />
    </div>
  );
}
