import { catalogIndex, type CatalogIndex } from "@/lib/catalog-index";

/**
 * useCatalog is the catalog of the behaviors, from the registry: each
 * behavior with its methods, the sections of the Behaviors page, and each
 * step that a plan can run. Each page reads the catalog through it, so a
 * catalog that the controller serves later changes only this hook.
 */
export function useCatalog(): CatalogIndex {
  return catalogIndex;
}
