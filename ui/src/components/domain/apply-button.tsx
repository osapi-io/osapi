import { Button } from "@/components/ui/button";
import { Kbd } from "@/components/ui/kbd";
import { Text } from "@/components/ui/text";
import { Play, Loader2, RotateCcw, X } from "lucide-react";

interface ApplyButtonProps {
  /** What the stack is, or after Apply what the run did: 12 applied: 11 ok. */
  summary: string;
  ready: boolean;
  applying: boolean;
  hasApplied: boolean;
  onApply: () => void;
  onReset: () => void;
  onClear: () => void;
}

/**
 * ApplyButton is the bar under the stack, pinned to the bottom of the
 * window, so Apply is in reach however long the stack is. After Apply it
 * keeps the stack and its results, says what the run did, and offers Clear
 * to start over and Run again, which makes the stack ready to apply again.
 */
export function ApplyButton({
  summary,
  ready,
  applying,
  hasApplied,
  onApply,
  onReset,
  onClear,
}: ApplyButtonProps) {
  return (
    <div className="sticky bottom-0 z-20 mt-6 flex items-center justify-end gap-3 border-t border-border bg-background/90 py-3 backdrop-blur-xl">
      <Text variant="muted" size="sm" className="mr-auto">
        {summary}
      </Text>
      {hasApplied && (
        <Button variant="ghost" size="md" onClick={onClear}>
          <X className="h-4 w-4" />
          Clear
        </Button>
      )}
      {hasApplied && (
        <Button variant="ghost" size="md" onClick={onReset}>
          <RotateCcw className="h-4 w-4" />
          Run again
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
