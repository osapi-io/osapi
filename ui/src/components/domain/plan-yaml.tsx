import { useState } from "react";
import { Check, Copy } from "lucide-react";
import { Modal } from "@/components/ui/modal";
import { Button } from "@/components/ui/button";
import { CodeBlock } from "@/components/ui/code-block";
import { Text } from "@/components/ui/text";

/**
 * PlanYAML shows a plan as a workflow file, read only, with a button that
 * copies it. A plan is not saved yet; the YAML is how to keep one.
 */
export function PlanYAML({
  open,
  onClose,
  yaml,
}: {
  open: boolean;
  onClose: () => void;
  yaml: string;
}) {
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    await navigator.clipboard.writeText(yaml);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };
  return (
    <Modal open={open} onClose={onClose} className="max-w-2xl">
      <h2 className="text-lg font-semibold text-white">The plan as YAML</h2>
      <Text variant="muted" as="p" className="mb-3 mt-1">
        The steps in the shape of a workflow file. The UI does not save a plan,
        so copy it to keep it.
      </Text>
      <CodeBlock variant="json" maxHeight="60vh">
        {yaml}
      </CodeBlock>
      <div className="mt-4 flex justify-end">
        <Button size="sm" onClick={copy}>
          {copied ? (
            <Check className="h-3.5 w-3.5" />
          ) : (
            <Copy className="h-3.5 w-3.5" />
          )}
          {copied ? "Copied" : "Copy"}
        </Button>
      </div>
    </Modal>
  );
}
