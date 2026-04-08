import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Modal } from "@/components/ui/modal";
import { Layers } from "lucide-react";

interface SaveStackDialogProps {
  open: boolean;
  blockCount: number;
  onSave: (name: string, description: string) => void;
  onClose: () => void;
}

export function SaveStackDialog({
  open,
  blockCount,
  onSave,
  onClose,
}: SaveStackDialogProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  const canSave = name.trim().length > 0;

  const handleSave = () => {
    if (canSave) {
      onSave(name.trim(), description.trim());
      setName("");
      setDescription("");
    }
  };

  return (
    <Modal open={open} onClose={onClose}>
      <div className="mb-4 flex items-center gap-2">
        <Layers className="h-5 w-5 text-primary" />
        <h2 className="text-lg font-semibold text-text">Save Stack</h2>
      </div>

      <p className="mb-4 text-sm text-text-muted">
        Save this configuration as a reusable stack ({blockCount} block
        {blockCount !== 1 ? "s" : ""}).
      </p>

      <div className="space-y-3">
        <Input
          id="stack-name"
          label="Name"
          placeholder="Deploy Nginx Config"
          value={name}
          onChange={(e) => setName(e.target.value)}
          autoFocus
        />
        <Input
          id="stack-description"
          label="Description (optional)"
          placeholder="Upload and deploy nginx.conf to all web servers"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
      </div>

      <div className="mt-6 flex justify-end gap-3">
        <Button variant="ghost" size="md" onClick={onClose}>
          Cancel
        </Button>
        <Button
          variant="primary"
          size="md"
          disabled={!canSave}
          onClick={handleSave}
        >
          Save
        </Button>
      </div>
    </Modal>
  );
}
