import { listPKITLS } from "@/sdk/gen/pki-api-pki-operations/pki-api-pki-operations";
import type { PKITLSSection, PKITLSList } from "@/sdk/gen/schemas";
import { useAuth } from "@/hooks/use-auth";
import { unwrap, usePolling } from "@/hooks/use-polling";

/** The TLS section of each component, each intervalMs. */
export function useTLS(intervalMs = 60000) {
  const { canCall } = useAuth();
  const { data, error } = usePolling(
    async () =>
      unwrap<PKITLSList>(
        await listPKITLS(),
        "Failed to fetch the TLS of each component",
      ).items,
    [],
    { enabled: canCall("ListPKITLS"), intervalMs },
  );
  const sections = data ?? [];
  /** The section of the component on the host, if it has one. */
  const find = (
    component?: string,
    hostname?: string,
  ): PKITLSSection | undefined =>
    sections.find((s) => s.component === component && s.hostname === hostname);
  return { sections, find, error };
}
