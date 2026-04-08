import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import { Dropdown } from "@/components/ui/dropdown";
import { ObjectPicker } from "@/components/domain/object-picker";
import type { BlockStatus } from "@/hooks/use-stack";

interface FileBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
  upstreamObjects?: string[];
}

export function FileBlock({
  data,
  onChange,
  onStatusChange,
  upstreamObjects = [],
}: FileBlockProps) {
  const objectName = (data.object_name as string) || "";
  const path = (data.path as string) || "";
  const mode = (data.mode as string) || "0644";
  const owner = (data.owner as string) || "";
  const group = (data.group as string) || "";
  const contentType = (data.content_type as string) || "raw";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const isReady = objectName.trim() !== "" && path.trim() !== "";
    const next: BlockStatus = isReady ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [objectName, path, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-3">
        <ObjectPicker
          id="deploy-object-name"
          label="Object"
          value={objectName}
          onChange={(v) => update("object_name", v)}
          upstreamObjects={upstreamObjects}
        />
        <Input
          id="deploy-path"
          label="Destination Path"
          placeholder="/etc/myapp/config.yaml"
          value={path}
          onChange={(e) => update("path", e.target.value)}
        />
      </div>
      <div className="grid grid-cols-4 gap-3">
        <Input
          id="deploy-mode"
          label="Mode"
          placeholder="0644"
          value={mode}
          onChange={(e) => update("mode", e.target.value)}
        />
        <Input
          id="deploy-owner"
          label="Owner"
          placeholder="root"
          value={owner}
          onChange={(e) => update("owner", e.target.value)}
        />
        <Input
          id="deploy-group"
          label="Group"
          placeholder="root"
          value={group}
          onChange={(e) => update("group", e.target.value)}
        />
        <Dropdown
          id="deploy-content-type"
          label="Type"
          value={contentType}
          onChange={(v) => update("content_type", v)}
          options={[
            { value: "raw", label: "Raw" },
            { value: "template", label: "Template" },
          ]}
        />
      </div>
    </div>
  );
}
