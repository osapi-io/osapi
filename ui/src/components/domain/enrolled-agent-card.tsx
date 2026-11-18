import { useState } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CopyField } from "@/components/ui/copy-field";
import { Text } from "@/components/ui/text";
import { Loader2, ShieldCheck, Trash2 } from "lucide-react";
import type { AgentInfo } from "@/sdk/gen/schemas";
import { deleteAgent } from "@/sdk/gen/agent-management-api-agent-operations/agent-management-api-agent-operations";

interface EnrolledAgentCardProps {
  agent: AgentInfo;
  onRefresh?: () => void;
}

// EnrolledAgentCard shows an agent whose key the controller stores, with a
// Remove action. Removing takes two clicks, because it stops the controller
// trusting the agent at once. The machine ID is always sent, so a hostname two
// machines share removes the one on this card.
export function EnrolledAgentCard({
  agent,
  onRefresh,
}: EnrolledAgentCardProps) {
  const [confirming, setConfirming] = useState(false);
  const [acting, setActing] = useState(false);
  const [removed, setRemoved] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleRemove = async () => {
    if (!confirming) {
      setConfirming(true);
      return;
    }
    setActing(true);
    setError(null);
    try {
      const resp = await deleteAgent(agent.hostname, {
        machine_id: agent.machine_id,
      });
      if (resp.status === 204) {
        setRemoved(true);
        setTimeout(() => onRefresh?.(), 1000);
      } else {
        setError((resp.data as { error?: string })?.error ?? "Remove failed");
      }
    } catch {
      setError("Remove failed");
    }
    setActing(false);
    setConfirming(false);
  };

  if (removed) {
    return (
      <Card variant="active" className="flex flex-col opacity-60">
        <CardContent className="flex items-center justify-center py-4">
          <Badge variant="muted">Removed</Badge>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card variant="active" className="flex flex-col">
      <CardHeader>
        <ShieldCheck className="h-4 w-4 text-primary" />
        <CardTitle className="truncate" title={agent.hostname}>
          {agent.hostname}
        </CardTitle>
        <div className="ml-auto">
          <Badge variant={agent.verified ? "ready" : "error"}>
            {agent.verified ? "Verified" : "Unverified"}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-1.5">
        <div className="space-y-0.5">
          {agent.machine_id && (
            <CopyField label="Machine ID" value={agent.machine_id} />
          )}
        </div>
        {error && (
          <Text variant="muted" as="p" className="text-status-error">
            {error}
          </Text>
        )}

        <div className="mt-auto flex gap-2 border-t border-border/30 pt-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleRemove}
            onBlur={() => setConfirming(false)}
            disabled={acting}
            className={confirming ? "text-status-error" : undefined}
          >
            {acting ? (
              <Loader2 className="h-3 w-3 animate-spin" />
            ) : (
              <Trash2 className="h-3 w-3" />
            )}
            {confirming ? "Confirm remove" : "Remove"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
