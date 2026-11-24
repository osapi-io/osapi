import { useEffect } from "react";
import { useLocation, useParams } from "react-router-dom";
import { ContentArea } from "@/components/layout/content-area";
import { Card } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { KVList } from "@/components/ui/kv-list";
import { SectionLabel } from "@/components/ui/section-label";
import { Text } from "@/components/ui/text";
import { useCatalog } from "@/hooks/use-catalog";
import {
  MethodCard,
  BehaviorLinks,
  Platforms,
} from "@/components/domain/method-card";
import { BackLink, PlanLink } from "@/components/domain/behavior-nav";

/**
 * BehaviorPage is one behavior, as osapi dsl behavior get and chain get print
 * it: what it manages, its area, platforms, and the behaviors that it calls
 * and that call it, then each method as a signature, with its permission
 * against the viewer's role, its platforms, its chain and its examples.
 */
export function BehaviorPage() {
  const { name = "" } = useParams();
  const { behavior, areas } = useCatalog();
  const b = behavior(name);
  const { hash } = useLocation();

  // A link to a method, /behaviors/dns#update, brings its card into view.
  useEffect(() => {
    if (!hash) return;
    document
      .getElementById(`method-${hash.slice(1)}`)
      ?.scrollIntoView({ block: "start" });
  }, [hash, name]);

  if (!b) {
    return (
      <ContentArea>
        <BackLink />
        <EmptyState message={`There is no behavior named "${name}".`} />
      </ContentArea>
    );
  }
  const area = areas.find((a) => a.name === b.area);

  return (
    <ContentArea>
      <BackLink area={b.internal ? "Internal" : area?.label} />
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div className="min-w-0">
          <div className="flex items-baseline gap-3">
            <h1 className="text-2xl font-bold text-white">{b.title}</h1>
            <span className="font-mono text-sm text-text-muted">{b.name}</span>
          </div>
          <p className="mt-1 max-w-[70ch] text-sm text-text-muted">{b.doc}</p>
        </div>
        <PlanLink />
      </div>

      <Card className="mb-8">
        <KVList
          className="gap-y-1.5"
          pairs={[
            ["Area", b.internal ? "internal" : (area?.label ?? b.area)],
            ["Under", b.parent ? <BehaviorLinks names={[b.parent]} /> : null],
            ["Platforms", <Platforms names={b.platforms} />],
            ["Calls", <BehaviorLinks names={b.uses} />],
            ["Called by", <BehaviorLinks names={b.used_by} />],
            [
              "Methods",
              `${b.methods.length}: ${b.methods.map((m) => m.name).join(", ")}`,
            ],
          ]}
        />
        {b.internal && (
          <Text variant="muted" as="p" className="mt-3">
            Only other behaviors call it. It has no route, so a plan cannot run
            it.
          </Text>
        )}
      </Card>

      <SectionLabel>Methods</SectionLabel>
      <div className="space-y-4">
        {b.methods.map((m) => (
          <MethodCard key={m.name} behavior={b} method={m} />
        ))}
      </div>
    </ContentArea>
  );
}
