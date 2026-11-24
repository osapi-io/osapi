import type { ReactNode } from "react";
import { Loader2, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Modal } from "@/components/ui/modal";
import { Text } from "@/components/ui/text";

interface ConfirmModalProps {
  open: boolean;
  title: string;
  /** What the person confirms, in a sentence. */
  children: ReactNode;
  /** The words of the button that does it, such as Delete. */
  action: string;
  busy?: boolean;
  onConfirm: () => void;
  onClose: () => void;
}

/** A question before a change that cannot be undone, such as a delete. */
export function ConfirmModal({
  open,
  title,
  children,
  action,
  busy,
  onConfirm,
  onClose,
}: ConfirmModalProps) {
  return (
    <Modal open={open} onClose={onClose}>
      <div className="space-y-4">
        <Text as="h3" className="text-lg font-semibold">
          {title}
        </Text>
        <Text variant="muted" as="p">
          {children}
        </Text>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="sm" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            size="sm"
            onClick={onConfirm}
            disabled={busy}
          >
            {busy ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <Trash2 className="h-3.5 w-3.5" />
            )}
            {action}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
