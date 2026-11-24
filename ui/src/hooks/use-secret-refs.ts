import { useEffect, useState } from "react";
import {
  getSecretVaults,
  getSecrets,
} from "@/sdk/gen/secret-management-api-secret-operations/secret-management-api-secret-operations";
import type { SecretListResponse, VaultListResponse } from "@/sdk/gen/schemas";
import { useAuth } from "@/hooks/use-auth";
import { secretRef } from "@/lib/secret-ref";

/** One reference that a secret field can take, and the vault it is in. */
export interface SecretRef {
  ref: string;
  vault: string;
  key: string;
}

/**
 * The references of each key of each vault, for the picker of a secret
 * field. It loads nothing for a role that may not list the vaults and
 * their keys, and a load that fails leaves the list empty: the field still
 * takes a reference that a person types.
 */
export function useSecretRefs() {
  const { canCall } = useAuth();
  const allowed = canCall("GetSecretVaults") && canCall("GetSecrets");
  const [refs, setRefs] = useState<SecretRef[]>([]);

  useEffect(() => {
    if (!allowed) return;
    let mounted = true;
    const load = async () => {
      try {
        const v = await getSecretVaults();
        if (v.status !== 200) return;
        const out: SecretRef[] = [];
        for (const vault of (v.data as VaultListResponse).items ?? []) {
          const k = await getSecrets(vault.name);
          if (k.status !== 200) continue;
          for (const s of (k.data as SecretListResponse).items ?? []) {
            out.push({
              ref: secretRef(vault.name, s.key),
              vault: vault.name,
              key: s.key,
            });
          }
        }
        if (mounted) setRefs(out);
      } catch {
        // The list stays empty.
      }
    };
    load();
    return () => {
      mounted = false;
    };
  }, [allowed]);

  return { refs, allowed };
}
