import { getFactKeys } from "@/sdk/gen/facts-api-facts/facts-api-facts";
import type { FactKeyEntry, FactKeysResponse } from "@/sdk/gen/schemas";
import { useLoad } from "@/hooks/use-load";

export type { FactKeyEntry };

const load = async () => {
  const r = await getFactKeys();
  return r.status === 200 ? (r.data as FactKeysResponse).items : undefined;
};

/** The fact keys the agents report, for suggestions. */
export function useFacts() {
  const { value } = useLoad<FactKeyEntry[]>(load, []);
  return { facts: value };
}
