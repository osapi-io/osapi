import { useState } from "react";
import { Card } from "@/components/ui/card";
import { Code, Copy, Check } from "lucide-react";
import { JobDetail } from "@/components/domain/job-detail";
import { CodeBlock } from "@/components/ui/code-block";
import { Text } from "@/components/ui/text";
import { CollapsibleSection } from "@/components/ui/collapsible-section";
import { ResultSummary } from "@/components/domain/result-summary";
import { hostOutcomes } from "@/lib/outcome";
import { resultsOf } from "@/lib/result-format";

/** HostError is the error of one host, under its row. */
export function HostError({ children }: { children: React.ReactNode }) {
  return (
    <Text variant="error" as="p" className="px-4 pb-2">
      {children}
    </Text>
  );
}

export function ChangedLabel({ changed }: { changed?: boolean }) {
  if (changed === undefined || changed === null) return null;
  return (
    <Text variant={changed ? "primary" : "label"}>
      {changed ? "changed" : "unchanged"}
    </Text>
  );
}

export function RawJsonToggle({ data }: { data: unknown }) {
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  const json = JSON.stringify(data, null, 2);

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation();
    await navigator.clipboard.writeText(json);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <CollapsibleSection
      open={open}
      onToggle={() => setOpen(!open)}
      icon={Code}
      label="JSON"
      rightContent={
        open ? (
          <span
            onClick={handleCopy}
            className="flex items-center gap-1 rounded px-1.5 py-0.5 hover:bg-white/5"
          >
            {copied ? (
              <Check className="h-3 w-3 text-primary" />
            ) : (
              <Copy className="h-3 w-3" />
            )}
            {copied ? "Copied" : "Copy"}
          </span>
        ) : undefined
      }
    >
      <CodeBlock variant="json">{json}</CodeBlock>
    </CollapsibleSection>
  );
}

export function ResultShell({
  jobId,
  label,
  summary,
  children,
  data,
}: {
  jobId?: string;
  label: string;
  /** The summary line; for a result with per-host results, it is counted from them. */
  summary?: React.ReactNode;
  children?: React.ReactNode;
  data: unknown;
}) {
  return (
    <Card variant="inset">
      <div className="flex items-center gap-2 border-b border-border/40 px-4 py-2.5">
        {summary ??
          (resultsOf(data).length > 0 && (
            <ResultSummary outcomes={hostOutcomes(resultsOf(data))} />
          ))}
        <Text variant="muted" className="font-semibold">
          {label}
        </Text>
        {jobId && (
          <div className="ml-auto">
            <JobDetail jobId={jobId} inline />
          </div>
        )}
      </div>
      {jobId && <JobDetail jobId={jobId} panel />}
      {children}
      <RawJsonToggle data={data} />
    </Card>
  );
}
