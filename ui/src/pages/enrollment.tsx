import { useCallback, useEffect, useState } from "react";
import { ContentArea } from "@/components/layout/content-area";
import { PendingAgentCard } from "@/components/domain/pending-agent-card";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { Text } from "@/components/ui/text";
import { useCommands } from "@/hooks/use-commands";
import { ShieldAlert } from "lucide-react";
import type { PendingAgentInfo } from "@/sdk/gen/schemas";
import {
  getAgentsPending,
  acceptAgent,
  rejectAgent,
} from "@/sdk/gen/agent-management-api-agent-operations/agent-management-api-agent-operations";

export function Enrollment() {
  const [pendingAgents, setPendingAgents] = useState<PendingAgentInfo[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchPending = useCallback(async () => {
    try {
      const resp = await getAgentsPending();
      if (resp.status === 200) {
        setPendingAgents(
          (resp.data as { agents?: PendingAgentInfo[] })?.agents ?? [],
        );
      }
    } catch {
      // silent
    }
    setLoading(false);
  }, []);

  useEffect(() => {
    fetchPending();
    const interval = setInterval(fetchPending, 10000);
    return () => clearInterval(interval);
  }, [fetchPending]);

  useCommands(
    [
      ...pendingAgents.map((a) => ({
        id: `pki:accept:${a.hostname}`,
        name: `accept ${a.hostname}`,
        description: `Accept PKI enrollment for ${a.hostname}`,
        category: "actions",
        action: async () => {
          await acceptAgent(a.hostname);
          fetchPending();
        },
      })),
      ...pendingAgents.map((a) => ({
        id: `pki:reject:${a.hostname}`,
        name: `reject ${a.hostname}`,
        description: `Reject PKI enrollment for ${a.hostname}`,
        category: "actions",
        action: async () => {
          await rejectAgent(a.hostname);
          fetchPending();
        },
      })),
    ],
    [pendingAgents, fetchPending],
  );

  return (
    <ContentArea>
      <PageHeader
        title="PKI Enrollment"
        subtitle="Manage agent PKI enrollment requests"
      />

      {!loading && pendingAgents.length === 0 && (
        <EmptyState
          icon={<ShieldAlert className="h-8 w-8 text-text-muted" />}
          message="No pending enrollment requests"
        />
      )}

      {pendingAgents.length > 0 && (
        <>
          <Text variant="muted" as="p" className="mb-4">
            {pendingAgents.length} agent{pendingAgents.length > 1 ? "s" : ""}{" "}
            awaiting PKI enrollment acceptance.
          </Text>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {pendingAgents.map((pa) => (
              <PendingAgentCard
                key={pa.machine_id}
                agent={pa}
                onRefresh={fetchPending}
              />
            ))}
          </div>
        </>
      )}
    </ContentArea>
  );
}
