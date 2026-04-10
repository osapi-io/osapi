import { useEffect, useRef, useState } from "react";
import { Input } from "@/components/ui/input";
import { Dropdown } from "@/components/ui/dropdown";
import { Text } from "@/components/ui/text";
import type { BlockStatus } from "@/hooks/use-stack";
import { Upload } from "lucide-react";

interface FileUploadBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function FileUploadBlock({
  data,
  onChange,
  onStatusChange,
}: FileUploadBlockProps) {
  const name = (data.name as string) || "";
  const contentType = (data.content_type as string) || "raw";
  const fileName = (data._file_name as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [dragOver, setDragOver] = useState(false);

  useEffect(() => {
    const isReady = name.trim() !== "" && data._file != null;
    const next: BlockStatus = isReady ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [name, data._file, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  const handleFile = (file: File) => {
    onChange({
      ...data,
      _file: file,
      _file_name: file.name,
      name: data.name || file.name,
    });
  };

  const handleFileInput = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) handleFile(file);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
    const file = e.dataTransfer.files[0];
    if (file) handleFile(file);
  };

  return (
    <div className="space-y-3">
      <div
        onDragOver={(e) => {
          e.preventDefault();
          setDragOver(true);
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={handleDrop}
        onClick={() => fileInputRef.current?.click()}
        className={`flex cursor-pointer items-center justify-center gap-2 rounded-md border-2 border-dashed px-4 py-3 transition-colors ${
          dragOver
            ? "border-primary/60 bg-primary/5"
            : fileName
              ? "border-primary/30 bg-primary/5"
              : "border-border hover:border-primary/30"
        }`}
      >
        <Upload className="h-4 w-4 text-text-muted" />
        <Text variant="muted">
          {fileName || "Drop a file or click to select"}
        </Text>
        <input
          ref={fileInputRef}
          type="file"
          className="hidden"
          onChange={handleFileInput}
        />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <Input
          id="upload-name"
          label="Object Name"
          placeholder="my-config"
          value={name}
          onChange={(e) => update("name", e.target.value)}
        />
        <Dropdown
          id="upload-content-type"
          label="Content Type"
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
