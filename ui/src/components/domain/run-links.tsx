import { useState } from "react";
import { Link } from "react-router-dom";
import { Check, Copy } from "lucide-react";
import { IdBadge } from "@/components/ui/id-badge";

interface RunLinksProps {
  /** The job that the call queued, when it queued one. */
  jobId?: string;
  /** The trace of the call, from the answer's traceresponse header. */
  traceId?: string;
  className?: string;
}

/**
 * RunLinks shows the job and the trace of one call of the API, short: a
 * click opens the job on the Jobs page or the trace on the Traces page, and
 * the copy button copies the whole ID.
 */
export function RunLinks({ jobId, traceId, className }: RunLinksProps) {
  if (!jobId && !traceId) return null;
  return (
    <div className={`flex flex-wrap items-center gap-2 ${className ?? ""}`}>
      {jobId && <RunLink label="job" id={jobId} to={`/admin/jobs/${jobId}`} />}
      {traceId && (
        <RunLink label="trace" id={traceId} to={`/admin/traces/${traceId}`} />
      )}
    </div>
  );
}

function RunLink({ label, id, to }: { label: string; id: string; to: string }) {
  const [copied, setCopied] = useState(false);
  const copy = () => {
    navigator.clipboard.writeText(id);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };
  return (
    <IdBadge>
      <Link
        to={to}
        title={`Open ${label} ${id}`}
        className="text-text-muted transition-colors hover:text-accent-light"
      >
        {label} {id.slice(0, 8)}
      </Link>
      <button
        type="button"
        onClick={copy}
        title={`Copy the ${label} ID`}
        aria-label={`Copy the ${label} ID`}
        className="text-text-muted/50 transition-colors hover:text-text-muted"
      >
        {copied ? (
          <Check className="h-3 w-3 text-primary" />
        ) : (
          <Copy className="h-3 w-3" />
        )}
      </button>
    </IdBadge>
  );
}
