import { useRef, useState } from "react";
import { enumLabel } from "@/lib/format";
import { PostFileBodyContentType } from "@/sdk/gen/schemas";
import { Input } from "@/components/ui/input";
import { Dropdown } from "@/components/ui/dropdown";
import { Text } from "@/components/ui/text";
import { Upload } from "lucide-react";

interface ObjectUploadFormProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
}

/**
 * ObjectUploadForm is the form of an upload to the object store: the file,
 * the name to keep it under, and whether it is raw or a template. The step
 * is ready when it has a file and a name.
 */
export function ObjectUploadForm({ data, onChange }: ObjectUploadFormProps) {
  const name = (data.name as string) || "";
  const contentType =
    (data.content_type as string) || PostFileBodyContentType.raw;
  const fileName = (data._file_name as string) || "";
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [dragOver, setDragOver] = useState(false);

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
          label="Name"
          placeholder="my-config"
          value={name}
          onChange={(e) => update("name", e.target.value)}
        />
        <Dropdown
          id="upload-content-type"
          label="Content Type"
          value={contentType}
          onChange={(v) => update("content_type", v)}
          options={Object.values(PostFileBodyContentType).map((v) => ({
            value: v,
            label: enumLabel(v),
          }))}
        />
      </div>
    </div>
  );
}
