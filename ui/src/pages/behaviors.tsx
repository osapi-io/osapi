import { useCallback, useMemo, useState } from "react";
import { ContentArea } from "@/components/layout/content-area";
import { PageHeader } from "@/components/ui/page-header";
import { SearchBox } from "@/components/ui/search-box";
import { EmptyState } from "@/components/ui/empty-state";
import { KindLabel } from "@/components/domain/method-kind";
import { BehaviorList } from "@/components/domain/behavior-list";
import { useCatalog } from "@/hooks/use-catalog";
import { useAuth } from "@/hooks/use-auth";
import { behaviorMatches } from "@/lib/catalog-index";
import { permitted, type Permission } from "@/lib/permissions";

/**
 * Behaviors is the catalog: what osapi manages on a host, by area. Each
 * behavior opens its page, with its methods, their permissions, platforms,
 * chains and examples. A read-only viewer can use all of it.
 */
export function Behaviors() {
  const { sections, internal } = useCatalog();
  const { can } = useAuth();
  const [query, setQuery] = useState("");
  const allowed = useCallback(
    (permission: Permission | undefined) => permitted(permission, can),
    [can],
  );

  const shown = useMemo(
    () =>
      sections
        .map((s) => ({
          ...s,
          behaviors: s.behaviors.filter((b) => behaviorMatches(b, query)),
        }))
        .filter((s) => s.behaviors.length > 0),
    [sections, query],
  );
  const shownInternal = internal.filter((b) => behaviorMatches(b, query));

  return (
    <ContentArea>
      <PageHeader
        title="Behaviors"
        subtitle="What osapi manages on a host, by area. Open a behavior to see its methods, then add one to a plan."
      />

      <div className="mb-6 flex flex-wrap items-center gap-x-6 gap-y-3">
        <SearchBox
          value={query}
          onChange={setQuery}
          onClose={() => setQuery("")}
          placeholder="Filter by behavior or method, such as dns update"
          autoFocus={false}
          className="w-full max-w-md"
          inputClassName="min-w-0 flex-1"
        />
        <div className="flex items-center gap-4">
          <KindLabel kind="read" />
          <KindLabel kind="write" />
          <KindLabel kind="action" />
        </div>
      </div>

      {shown.length === 0 && shownInternal.length === 0 ? (
        <EmptyState message={`No behavior matches "${query}".`} />
      ) : (
        <div className="space-y-8">
          {shown.map((s) => (
            <BehaviorList
              key={s.area}
              label={s.label}
              behaviors={s.behaviors}
              allowed={allowed}
            />
          ))}
          {shownInternal.length > 0 && (
            <BehaviorList
              label="Internal"
              note="Only other behaviors call these. They have no route, so a plan cannot run them."
              behaviors={shownInternal}
              allowed={allowed}
            />
          )}
        </div>
      )}
    </ContentArea>
  );
}
