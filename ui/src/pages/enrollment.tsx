import { useCallback, useEffect, useState } from "react";
import { ContentArea } from "@/components/layout/content-area";
import { PendingAgentCard } from "@/components/domain/pending-agent-card";
import { EnrolledAgentCard } from "@/components/domain/enrolled-agent-card";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { Text } from "@/components/ui/text";
import { useCommands } from "@/hooks/use-commands";
import { useAuth } from "@/hooks/use-auth";
import { ErrorBanner } from "@/components/ui/error-banner";
import { Spinner } from "@/components/ui/spinner";
import { PermissionDenied } from "@/components/layout/permission-denied";
import { ShieldAlert } from "lucide-react";
import type {
  AgentInfo,
  ListAgentsResponse,
  PendingAgentInfo,
} from "@/sdk/gen/schemas";
import {
  getAgents,
  getAgentsPending,
  acceptAgent,
  rejectAgent,
  deleteAgent,
} from "@/sdk/gen/agent-management-api-agent-operations/agent-management-api-agent-operations";

export function Enrollment() {
  const { can } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [pendingAgents, setPendingAgents] = useState<PendingAgentInfo[]>([]);
  const [enrolledAgents, setEnrolledAgents] = useState<AgentInfo[]>([]);
  const [loading, setLoading] = useState(true);

  // A failed load shows as an error, as on the other pages, not as an
  // empty page that reads as no agents.
  const fetchPending = useCallback(async () => {
    const failed = (what: string, e?: unknown) =>
      e instanceof Error ? e.message : `Failed to fetch ${what}`;
    let err: string | null = null;
    try {
      const resp = await getAgentsPending();
      if (resp.status === 200) {
        setPendingAgents(
          (resp.data as { items?: PendingAgentInfo[] })?.items ?? [],
        );
      } else {
        err = failed("pending agents");
      }
    } catch (e) {
      err = failed("pending agents", e);
    }
    try {
      // An agent whose key the controller stores is enrolled.
      const resp = await getAgents();
      if (resp.status === 200) {
        setEnrolledAgents(
          ((resp.data as ListAgentsResponse).items ?? []).filter(
            (a) => a.key_stored,
          ),
        );
      } else {
        err ??= failed("agents");
      }
    } catch (e) {
      err ??= failed("agents", e);
    }
    setError(err);
    setLoading(false);
  }, []);

  useEffect(() => {
    const initialFetch = setTimeout(fetchPending, 0);
    const interval = setInterval(fetchPending, 10000);
    return () => {
      clearTimeout(initialFetch);
      clearInterval(interval);
    };
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
      ...enrolledAgents.map((a) => ({
        id: `pki:remove:${a.machine_id ?? a.hostname}`,
        name: `remove ${a.hostname}`,
        description: `Remove the enrolled agent ${a.hostname}; its key stops verifying`,
        category: "actions",
        action: async () => {
          await deleteAgent(a.hostname, { machine_id: a.machine_id });
          fetchPending();
        },
      })),
    ],
    [pendingAgents, enrolledAgents, fetchPending],
  );

  if (!can("agent:read")) {
    return <PermissionDenied title="PKI Enrollment" what="agents" />;
  }

  return (
    <ContentArea>
      <PageHeader
        title="PKI Enrollment"
        subtitle="Accept, reject and remove agents' PKI enrollment"
      />

      {loading && <Spinner />}

      {error && <ErrorBanner message={error} />}

      {!loading && !error && pendingAgents.length === 0 && (
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

      {enrolledAgents.length > 0 && (
        <>
          <Text variant="muted" as="p" className="mb-4 mt-8">
            {enrolledAgents.length} enrolled agent
            {enrolledAgents.length > 1 ? "s" : ""}. Removing one stops the
            controller trusting it at once; to enroll it again, delete
            controller.pub on the host and restart the agent.
          </Text>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {enrolledAgents.map((a) => (
              <EnrolledAgentCard
                key={a.machine_id ?? a.hostname}
                agent={a}
                onRefresh={fetchPending}
              />
            ))}
          </div>
        </>
      )}
    </ContentArea>
  );
}
