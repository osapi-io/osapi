import { useEffect, useState } from "react";
import { getFactKeys } from "@/sdk/gen/facts-api-facts/facts-api-facts";
import type { FactKeyEntry, FactKeysResponse } from "@/sdk/gen/schemas";

export type { FactKeyEntry };

export function useFacts() {
  const [facts, setFacts] = useState<FactKeyEntry[]>([]);

  useEffect(() => {
    let mounted = true;

    const fetch = async () => {
      try {
        const result = await getFactKeys();
        if (mounted && result.status === 200) {
          setFacts((result.data as FactKeysResponse).keys);
        }
      } catch {
        // silent — fact suggestions are non-critical
      }
    };

    fetch();
    return () => {
      mounted = false;
    };
  }, []);

  return { facts };
}
