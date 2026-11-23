import { Link } from "react-router-dom";
import { ChevronRight, CornerDownRight, Lock } from "lucide-react";
import { SectionLabel } from "@/components/ui/section-label";
import { Text } from "@/components/ui/text";
import { KindMark } from "@/components/domain/method-kind";
import { runName } from "@/lib/catalog-index";
import type { CatalogBehavior } from "@/lib/catalog";

interface BehaviorListProps {
  /** The heading of the list, such as an area's label. */
  label: string;
  /** A line under the heading. */
  note?: string;
  behaviors: CatalogBehavior[];
  /** Whether the viewer's role may run a method that needs the permission. */
  allowed: (permission: string | undefined) => boolean;
}

/**
 * BehaviorList is the behaviors of one area, each a row that opens its page:
 * its title and name, what it manages, a mark for each method (read, write
 * or action, faint when the viewer's role cannot run it), and the behaviors
 * that it calls. A behavior under a parent, such as the SSH keys of a user,
 * is under its parent's row.
 */
export function BehaviorList({
  label,
  note,
  behaviors,
  allowed,
}: BehaviorListProps) {
  return (
    <section>
      <div className="flex items-baseline gap-3">
        <SectionLabel>{label}</SectionLabel>
        <Text variant="muted" className="mb-2">
          {behaviors.length} {behaviors.length === 1 ? "behavior" : "behaviors"}
        </Text>
      </div>
      {note && (
        <Text variant="muted" as="p" className="-mt-1 mb-2">
          {note}
        </Text>
      )}
      <ul className="divide-y divide-border overflow-hidden rounded-lg border border-border bg-card">
        {behaviors.map((b) => (
          <BehaviorRow key={b.name} behavior={b} allowed={allowed} />
        ))}
      </ul>
    </section>
  );
}

function BehaviorRow({
  behavior: b,
  allowed,
}: {
  behavior: CatalogBehavior;
  allowed: (permission: string | undefined) => boolean;
}) {
  const locked = b.methods.map((m) => !b.internal && !allowed(m.permission));
  const none = !b.internal && locked.every(Boolean);
  return (
    <li>
      <Link
        to={`/behaviors/${b.name}`}
        className="group grid grid-cols-[minmax(0,17rem)_minmax(0,1fr)_auto] items-center gap-x-6 gap-y-1 px-4 py-3 transition-colors hover:bg-white/[0.03] max-md:grid-cols-[minmax(0,1fr)_auto]"
      >
        <div className="flex min-w-0 items-center gap-2">
          {b.parent && (
            <CornerDownRight className="h-3.5 w-3.5 shrink-0 text-text-muted/60" />
          )}
          <span className="truncate text-sm font-semibold text-white">
            {b.title}
          </span>
          <span className="truncate font-mono text-xs text-text-muted">
            {b.name}
          </span>
        </div>
        <Text
          variant="muted"
          size="sm"
          truncate
          className="max-md:col-span-2 max-md:row-start-2"
        >
          {b.doc}
        </Text>
        <div className="flex items-center gap-3">
          {b.uses.length > 0 && (
            <span
              className="hidden font-mono text-xs text-text-muted lg:inline"
              title={`Calls ${b.uses.join(", ")}`}
            >
              calls {b.uses.join(", ")}
            </span>
          )}
          <span className="flex items-center gap-0.5">
            {b.methods.map((m, i) => (
              <KindMark
                key={m.name}
                kind={m.kind}
                locked={locked[i]}
                title={`${runName(b.name, m.name)}: ${m.kind}${locked[i] ? `, needs ${m.permission ?? "a permission that is not known"}` : ""}`}
              />
            ))}
          </span>
          <span className="w-4">
            {none ? (
              <Lock
                className="h-3.5 w-3.5 text-text-muted/60"
                aria-label="Your role can run none of its methods"
              />
            ) : (
              <ChevronRight className="h-4 w-4 text-text-muted/40 transition-colors group-hover:text-text-muted" />
            )}
          </span>
        </div>
      </Link>
    </li>
  );
}
