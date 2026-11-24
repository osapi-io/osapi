import type { ComponentEntry } from "@/sdk/gen/schemas";

// SubRow is one row under a component card: a sub-component of the
// component, or a check that the controller made, such as its connection
// to NATS.
export type SubRow = {
  name: string;
  status: string;
  address?: string;
  url?: string;
};

// subRows are the sub-components of entry, as the API sorts them.
export function subRows(entry?: ComponentEntry): SubRow[] {
  return (entry?.components ?? []).map((c) => ({
    name: c.type ?? "",
    status: c.status ?? "",
    address: c.address,
    url: c.url,
  }));
}
