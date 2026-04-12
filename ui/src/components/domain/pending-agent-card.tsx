import { useState } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CopyField } from "@/components/ui/copy-field";
import { Text } from "@/components/ui/text";
import { Check, X, Loader2, ShieldAlert } from "lucide-react";
import type { PendingAgentInfo } from "@/sdk/gen/schemas";
import {
  acceptAgent,
  rejectAgent,
} from "@/sdk/gen/agent-management-api-agent-operations/agent-management-api-agent-operations";

interface PendingAgentCardProps {
  agent: PendingAgentInfo;
  onRefresh?: () => void;
}

function formatAge(ts?: string) {
  if (!ts) return "-";
  const ms = Date.now() - new Date(ts).getTime();
  const mins = Math.floor(ms / 60000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

export function PendingAgentCard({ agent, onRefresh }: PendingAgentCardProps) {
  const [acting, setActing] = useState(false);
  const [result, setResult] = useState<"accepted" | "rejected" | null>(null);

  const handleAccept = async () => {
    setActing(true);
    try {
      await acceptAgent(agent.hostname);
      setResult("accepted");
      setTimeout(() => onRefresh?.(), 1000);
    } catch {
      // silent — next poll will show state
    }
    setActing(false);
  };

  const handleReject = async () => {
    setActing(true);
    try {
      await rejectAgent(agent.hostname);
      setResult("rejected");
      setTimeout(() => onRefresh?.(), 1000);
    } catch {
      // silent
    }
    setActing(false);
  };

  if (result) {
    return (
      <Card variant="active" className="flex flex-col opacity-60">
        <CardContent className="flex items-center justify-center py-4">
          <Badge variant={result === "accepted" ? "ready" : "muted"}>
            {result === "accepted" ? "Accepted" : "Rejected"}
          </Badge>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card variant="active" className="flex flex-col">
      <CardHeader>
        <ShieldAlert className="h-4 w-4 text-accent-light" />
        <CardTitle className="truncate" title={agent.hostname}>
          {agent.hostname}
        </CardTitle>
        <div className="ml-auto">
          <Badge variant="pending">Pending</Badge>
        </div>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-1.5">
        <div className="space-y-0.5">
          <CopyField label="ID" value={agent.machine_id} />
          <CopyField label="FP" value={agent.fingerprint} />
        </div>
        <Text variant="muted" as="p">
          Requested {formatAge(agent.requested_at)}
        </Text>

        <div className="mt-auto flex gap-2 border-t border-border/30 pt-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleAccept}
            disabled={acting}
            className="text-primary"
          >
            {acting ? (
              <Loader2 className="h-3 w-3 animate-spin" />
            ) : (
              <Check className="h-3 w-3" />
            )}
            Accept
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleReject}
            disabled={acting}
          >
            {acting ? (
              <Loader2 className="h-3 w-3 animate-spin" />
            ) : (
              <X className="h-3 w-3" />
            )}
            Reject
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
