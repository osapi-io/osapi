import { Button } from "@/components/ui/button";
import { Kbd } from "@/components/ui/kbd";
import { Text } from "@/components/ui/text";
import { Play, Loader2, RotateCcw, X } from "lucide-react";

interface RunBarProps {
  /** What the plan is, or after Run what the run did: 3 ran: 3 ok. */
  summary: string;
  ready: boolean;
  running: boolean;
  hasRun: boolean;
  onRun: () => void;
  onReset: () => void;
  onClear: () => void;
}

/**
 * RunBar is the bar under the plan, pinned to the bottom of the window, so
 * Run is in reach however long the plan is. After a run it keeps the steps
 * and their results, says what the run did, and offers Clear to start over
 * and Run again, which makes the plan ready to run again.
 */
export function RunBar({
  summary,
  ready,
  running,
  hasRun,
  onRun,
  onReset,
  onClear,
}: RunBarProps) {
  return (
    <div className="sticky bottom-0 z-20 mt-6 flex items-center justify-end gap-3 border-t border-border bg-background/90 py-3 backdrop-blur-xl">
      <Text variant="muted" size="sm" className="mr-auto">
        {summary}
      </Text>
      {hasRun && (
        <Button variant="ghost" size="md" onClick={onClear}>
          <X className="h-4 w-4" />
          Clear
        </Button>
      )}
      {hasRun && (
        <Button variant="ghost" size="md" onClick={onReset}>
          <RotateCcw className="h-4 w-4" />
          Run again
        </Button>
      )}
      <Button
        variant="primary"
        size="lg"
        disabled={!ready || running}
        onClick={onRun}
      >
        {running ? (
          <>
            <Loader2 className="h-4 w-4 animate-spin" />
            Running...
          </>
        ) : (
          <>
            <Play className="h-4 w-4" />
            Run
            <Kbd className="ml-1">⌘↵</Kbd>
          </>
        )}
      </Button>
    </div>
  );
}
