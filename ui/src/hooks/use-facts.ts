import { getAgentFacts } from "@/sdk/gen/agent-management-api-agent-operations/agent-management-api-agent-operations";
import type { FactKeyEntry, FactKeysResponse } from "@/sdk/gen/schemas";
import { useLoad } from "@/hooks/use-load";

export type { FactKeyEntry };

const load = async () => {
  const r = await getAgentFacts();
  return r.status === 200 ? (r.data as FactKeysResponse).items : undefined;
};

/** The fact keys the agents report, for suggestions. */
export function useFacts() {
  const { value } = useLoad<FactKeyEntry[]>(load, []);
  return { facts: value };
}
