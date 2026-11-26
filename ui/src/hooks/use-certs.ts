import { listPKICerts } from "@/sdk/gen/pki-api-pki-operations/pki-api-pki-operations";
import type { PKICertList } from "@/sdk/gen/schemas";
import { useAuth } from "@/hooks/use-auth";
import { unwrap, usePolling } from "@/hooks/use-polling";

/** The certificates of the fleet, each intervalMs. */
export function useCerts(intervalMs = 60000) {
  const { canCall } = useAuth();
  const { data, error } = usePolling(
    async () =>
      unwrap<PKICertList>(
        await listPKICerts(),
        "Failed to fetch the certificates",
      ).items,
    [],
    { enabled: canCall("ListPKICerts"), intervalMs },
  );
  return { certs: data ?? [], error };
}
