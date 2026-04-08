import { cn } from "@/lib/cn";
import type { HTMLAttributes } from "react";

type CodeBlockVariant = "stdout" | "stderr" | "error" | "json" | "muted";

interface CodeBlockProps extends HTMLAttributes<HTMLPreElement> {
  variant?: CodeBlockVariant;
  maxHeight?: string;
}

const variantClasses: Record<CodeBlockVariant, string> = {
  stdout:
    "overflow-auto whitespace-pre-wrap rounded bg-[#0a0a0a] px-2 py-1.5 font-mono text-xs text-text/80",
  stderr:
    "overflow-auto whitespace-pre-wrap rounded bg-status-running/5 px-2 py-1.5 font-mono text-xs text-status-running",
  error:
    "overflow-auto whitespace-pre-wrap rounded bg-status-error/5 px-2 py-1.5 font-mono text-xs text-status-error",
  json: "overflow-auto px-4 pb-2 font-mono text-xs text-text/60",
  muted:
    "overflow-auto rounded bg-background/50 px-2 py-1 font-mono text-xs text-text/70",
};

const defaultMaxHeight: Record<CodeBlockVariant, string> = {
  stdout: "max-h-48",
  stderr: "max-h-32",
  error: "",
  json: "max-h-40",
  muted: "max-h-24",
};

export function CodeBlock({
  variant = "stdout",
  maxHeight,
  className,
  children,
  ...props
}: CodeBlockProps) {
  return (
    <pre
      className={cn(
        variantClasses[variant],
        maxHeight ?? defaultMaxHeight[variant],
        className,
      )}
      {...props}
    >
      {children}
    </pre>
  );
}
