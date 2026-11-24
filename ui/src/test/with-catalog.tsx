import type { ReactNode } from "react";
import { CatalogContext } from "@/lib/catalog-context";
import { testIndex } from "@/test/catalog";

/** The test catalog, as CatalogProvider gives it to a component under test. */
export function WithCatalog({ children }: { children: ReactNode }) {
  return (
    <CatalogContext.Provider value={testIndex}>
      {children}
    </CatalogContext.Provider>
  );
}
