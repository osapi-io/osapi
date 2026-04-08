import { SingleInputBlock } from "./single-input-block";
import type { BlockStatus } from "@/hooks/use-stack";

interface ServiceActionBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function ServiceActionBlock(props: ServiceActionBlockProps) {
  return (
    <SingleInputBlock
      {...props}
      field="name"
      label="Service Name"
      placeholder="nginx.service"
    />
  );
}
