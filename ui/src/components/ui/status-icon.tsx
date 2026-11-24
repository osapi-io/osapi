import { CheckCircle, Clock, XCircle, MinusCircle } from "lucide-react";
import { RowStatus } from "@/sdk/gen/schemas";

interface StatusIconProps {
  status: RowStatus;
  className?: string;
}

export function StatusIcon({ status, className }: StatusIconProps) {
  const size = className ?? "h-3 w-3 shrink-0";
  switch (status) {
    case RowStatus.ok:
      return <CheckCircle className={`${size} text-primary`} />;
    case RowStatus.failed:
      return <XCircle className={`${size} text-status-error`} />;
    case RowStatus.skipped:
      return <MinusCircle className={`${size} text-text-muted`} />;
    // The host did not answer: not a failure, since nobody knows whether
    // the operation ran.
    case RowStatus.timeout:
      return <Clock className={`${size} text-status-error`} />;
  }
}
