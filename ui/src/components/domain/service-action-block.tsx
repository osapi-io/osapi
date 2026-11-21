import { SingleInputBlock } from "./single-input-block";
import type { BlockStatus } from "@/hooks/use-stack";

interface ServiceActionBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
  /** A read names any service; a write names one that osapi deployed. */
  anyService?: boolean;
}

export function ServiceActionBlock({
  anyService = false,
  ...props
}: ServiceActionBlockProps) {
  return (
    <SingleInputBlock
      {...props}
      field="name"
      label="Service Name"
      placeholder={anyService ? "ssh" : "app, for osapi-app.service"}
    />
  );
}
