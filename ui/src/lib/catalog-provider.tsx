import { useCallback, useEffect, useState, type ReactNode } from "react";
import { Button } from "@/components/ui/button";
import { ErrorBanner } from "@/components/ui/error-banner";
import { Spinner } from "@/components/ui/spinner";
import { useSession } from "@/hooks/use-session";
import { unwrap } from "@/hooks/use-polling";
import { CatalogContext } from "@/lib/catalog-context";
import { buildIndex, type CatalogIndex } from "@/lib/catalog-index";
import { getCatalog } from "@/sdk/gen/catalog-api-catalog/catalog-api-catalog";
import type { CatalogResponse } from "@/sdk/gen/schemas";

/**
 * CatalogProvider loads the catalog once after sign-in, from
 * GET /api/catalog, and shows its children only when it has it. Until then
 * it shows a spinner. When the load fails, such as a role without
 * catalog:read, it shows why, with a retry and a sign out.
 */
export function CatalogProvider({ children }: { children: ReactNode }) {
  const { clearToken } = useSession();
  const [index, setIndex] = useState<CatalogIndex | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let live = true;
    getCatalog()
      .then((res) => unwrap<CatalogResponse>(res, "Failed to load the catalog"))
      .then((catalog) => {
        if (live) setIndex(buildIndex(catalog));
      })
      .catch((err: unknown) => {
        if (live) {
          setError(err instanceof Error ? err.message : String(err));
        }
      });
    return () => {
      live = false;
    };
  }, [attempt]);

  const retry = useCallback(() => {
    setError(null);
    setAttempt((n) => n + 1);
  }, []);

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="w-full max-w-md space-y-4 px-6">
          <ErrorBanner message={`Could not load the catalog: ${error}`} />
          <div className="flex justify-end gap-2">
            <Button variant="secondary" onClick={clearToken}>
              Sign out
            </Button>
            <Button onClick={retry}>Retry</Button>
          </div>
        </div>
      </div>
    );
  }

  if (!index) {
    return (
      <div className="min-h-screen bg-background">
        <Spinner />
      </div>
    );
  }

  return (
    <CatalogContext.Provider value={index}>{children}</CatalogContext.Provider>
  );
}
