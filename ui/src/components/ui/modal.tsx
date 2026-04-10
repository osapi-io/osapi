import { cn } from "@/lib/cn";
import { X } from "lucide-react";
import { IconButton } from "@/components/ui/icon-button";
import type { ReactNode } from "react";

interface ModalProps {
  open: boolean;
  onClose: () => void;
  children: ReactNode;
  className?: string;
}

export function Modal({ open, onClose, children, className }: ModalProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="absolute inset-0 bg-black/60" onClick={onClose} />
      <div
        className={cn(
          "relative w-full max-w-md rounded-lg border border-border/60 bg-card p-6 shadow-xl",
          className,
        )}
      >
        <div className="absolute right-4 top-4">
          <IconButton icon={X} variant="ghost" onClick={onClose} />
        </div>
        {children}
      </div>
    </div>
  );
}
