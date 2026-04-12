import { useState } from "react";
import { Text } from "@/components/ui/text";
import { Copy, Check } from "lucide-react";

interface CopyFieldProps {
  label: string;
  value: string;
}

export function CopyField({ label, value }: CopyFieldProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div className="flex items-center gap-1">
      <Text variant="label" as="span" className="shrink-0 text-[10px]">
        {label}:
      </Text>
      <Text
        variant="muted"
        as="span"
        className="truncate font-mono text-[10px]"
        title={value}
      >
        {value}
      </Text>
      <button
        onClick={handleCopy}
        className="shrink-0 text-text-muted/50 transition-colors hover:text-text-muted"
        title="Copy to clipboard"
      >
        {copied ? (
          <Check className="h-2.5 w-2.5 text-primary" />
        ) : (
          <Copy className="h-2.5 w-2.5" />
        )}
      </button>
    </div>
  );
}
