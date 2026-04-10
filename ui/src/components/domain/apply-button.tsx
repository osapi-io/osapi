import { Button } from "@/components/ui/button";
import { Kbd } from "@/components/ui/kbd";
import { Play, Loader2, RotateCcw, Save } from "lucide-react";

interface ApplyButtonProps {
  ready: boolean;
  applying: boolean;
  hasApplied: boolean;
  hasBlocks: boolean;
  showSave?: boolean;
  onApply: () => void;
  onReset: () => void;
  onSave: () => void;
}

export function ApplyButton({
  ready,
  applying,
  hasApplied,
  hasBlocks,
  showSave,
  onApply,
  onReset,
  onSave,
}: ApplyButtonProps) {
  return (
    <div className="flex items-center justify-end gap-3 pt-6">
      {hasApplied && (
        <Button variant="ghost" size="md" onClick={onReset}>
          <RotateCcw className="h-4 w-4" />
          Reset
        </Button>
      )}
      {showSave && (
        <Button
          variant="ghost"
          size="md"
          disabled={!hasBlocks}
          onClick={onSave}
        >
          <Save className="h-4 w-4" />
          Save Stack
        </Button>
      )}
      <Button
        variant="primary"
        size="lg"
        disabled={!ready || applying}
        onClick={onApply}
      >
        {applying ? (
          <>
            <Loader2 className="h-4 w-4 animate-spin" />
            Applying...
          </>
        ) : (
          <>
            <Play className="h-4 w-4" />
            Apply
            <Kbd className="ml-1">⌘↵</Kbd>
          </>
        )}
      </Button>
    </div>
  );
}
