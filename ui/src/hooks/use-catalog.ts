import { useContext } from "react";
import type { CatalogIndex } from "@/lib/catalog-index";
import { CatalogContext } from "@/lib/catalog-context";

/**
 * useCatalog is the catalog that the controller serves at GET /api/catalog,
 * indexed: each behavior with its methods, the sections of the Behaviors
 * page, each step that a plan can run, the roles, the permission of each
 * operation, and the tones. CatalogProvider loads it once after sign-in and
 * shows the pages only after it loads, so a page always has it.
 */
export function useCatalog(): CatalogIndex {
  const ctx = useContext(CatalogContext);
  if (!ctx) throw new Error("useCatalog must be used within CatalogProvider");
  return ctx;
}
