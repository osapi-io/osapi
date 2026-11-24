import { useMemo } from "react";
import { useCatalog } from "@/hooks/use-catalog";
import { statusDot, statusText, statusVariant } from "@/lib/status";

/**
 * useStatus gives the colors of a status word with the tones of the
 * catalog: the Badge variant, the HealthDot color and the Text variant. A
 * page does not map a status to a color of its own.
 */
export function useStatus() {
  const { statusTones } = useCatalog();
  return useMemo(
    () => ({
      variant: (status?: string | null) => statusVariant(statusTones, status),
      dot: (status?: string | null) => statusDot(statusTones, status),
      text: (status?: string | null) => statusText(statusTones, status),
    }),
    [statusTones],
  );
}
