import { createContext } from "react";
import type { CatalogIndex } from "@/lib/catalog-index";

/** The catalog that the controller served after sign-in, indexed. */
export const CatalogContext = createContext<CatalogIndex | null>(null);
