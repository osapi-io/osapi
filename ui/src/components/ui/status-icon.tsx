import { CheckCircle, XCircle, MinusCircle } from "lucide-react";

type Status = "ok" | "failed" | "skipped";

interface StatusIconProps {
  status: Status;
  className?: string;
}

export function StatusIcon({ status, className }: StatusIconProps) {
  const size = className ?? "h-3 w-3 shrink-0";
  switch (status) {
    case "ok":
      return <CheckCircle className={`${size} text-primary`} />;
    case "failed":
      return <XCircle className={`${size} text-status-error`} />;
    case "skipped":
      return <MinusCircle className={`${size} text-text-muted`} />;
  }
}
