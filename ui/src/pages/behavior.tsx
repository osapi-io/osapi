import { useEffect, useState } from "react";
import { Link, useLocation, useParams } from "react-router-dom";
import { Check, ChevronLeft, ListPlus, Lock } from "lucide-react";
import { ContentArea } from "@/components/layout/content-area";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { CodeBlock } from "@/components/ui/code-block";
import { CollapsibleSection } from "@/components/ui/collapsible-section";
import { EmptyState } from "@/components/ui/empty-state";
import { KVList } from "@/components/ui/kv-list";
import { SectionLabel } from "@/components/ui/section-label";
import { Text } from "@/components/ui/text";
import { CallTree } from "@/components/domain/call-tree";
import { KindMark } from "@/components/domain/method-kind";
import { useAuth } from "@/hooks/use-auth";
import { useCatalog } from "@/hooks/use-catalog";
import { usePlan } from "@/hooks/use-plan";
import { AREAS } from "@/lib/catalog.gen";
import { runName } from "@/lib/catalog-index";
import {
  methodKindLabels,
  methodLabel,
  type CatalogBehavior,
  type CatalogMethod,
} from "@/lib/catalog";
import { permitted } from "@/lib/permissions";

/** A list of behaviors, each a link to its page, or "none". */
function BehaviorLinks({ names }: { names: readonly string[] }) {
  if (names.length === 0) return <span className="text-text-muted">none</span>;
  return (
    <span className="flex flex-wrap gap-x-3">
      {names.map((n) => (
        <Link
          key={n}
          to={`/behaviors/${n}`}
          className="font-mono text-heading-1 hover:underline"
        >
          {n}
        </Link>
      ))}
    </span>
  );
}

/** Platforms as the agent names them, each a tag. */
function Platforms({ names }: { names: readonly string[] }) {
  if (names.length === 0) {
    return <span className="text-text-muted">each platform</span>;
  }
  return (
    <span className="flex flex-wrap gap-1.5">
      {names.map((p) => (
        <span
          key={p}
          className="rounded border border-border px-1.5 py-px font-mono text-[11px] text-text"
        >
          {p}
        </span>
      ))}
    </span>
  );
}

/**
 * BehaviorPage is one behavior, as osapi dsl behavior get and chain get print
 * it: what it manages, its area, platforms, and the behaviors that it calls
 * and that call it, then each method as a signature, with its permission
 * against the viewer's role, its platforms, its chain and its examples.
 */
export function BehaviorPage() {
  const { name = "" } = useParams();
  const { behavior } = useCatalog();
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
  const area = AREAS.find((a) => a.name === b.area);

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
            ["Called by", <BehaviorLinks names={b.usedBy} />],
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

function BackLink({ area }: { area?: string }) {
  return (
    <Link
      to="/behaviors"
      className="mb-4 inline-flex items-center gap-1 text-sm text-text-muted transition-colors hover:text-text"
    >
      <ChevronLeft className="h-4 w-4" />
      Behaviors
      {area && <span className="text-text-muted/60">/ {area}</span>}
    </Link>
  );
}

/** The way to the plan, with its count of steps, once it has one. */
function PlanLink() {
  const { steps } = usePlan();
  if (steps.length === 0) return null;
  return (
    <Link
      to="/plans/new"
      className="inline-flex h-8 items-center gap-2 rounded-md border border-border px-3 text-xs font-semibold text-text transition-colors hover:border-white/20"
    >
      Open the plan: {steps.length} {steps.length === 1 ? "step" : "steps"}
    </Link>
  );
}

function MethodCard({
  behavior: b,
  method: m,
}: {
  behavior: CatalogBehavior;
  method: CatalogMethod;
}) {
  const { can, role } = useAuth();
  const { addStep } = usePlan();
  const { step } = useCatalog();
  const [examplesOpen, setExamplesOpen] = useState(false);
  const [added, setAdded] = useState(0);
  const allowed = permitted(m.permission, can);
  const run = runName(b.name, m.name);

  return (
    <Card id={`method-${m.name}`} className="scroll-mt-24">
      <div className="flex flex-wrap items-start gap-3">
        <KindMark kind={m.kind} className="mt-0.5" />
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-baseline gap-x-3">
            <h3 className="text-sm font-semibold text-white">
              {methodLabel(m.name)}
            </h3>
            <span className="font-mono text-xs text-text-muted">{run}</span>
            <span className="text-xs text-text-muted">
              {methodKindLabels[m.kind]}
            </span>
          </div>
          <Text as="p" size="sm" className="mt-0.5">
            {m.doc}
          </Text>
        </div>
        {!b.internal && (
          <div className="flex items-center gap-2">
            {added > 0 && (
              <Link
                to="/plans/new"
                className="inline-flex items-center gap-1 text-xs text-text-muted hover:text-text"
              >
                <Check className="h-3.5 w-3.5" />
                Added, open the plan
              </Link>
            )}
            <Button
              size="sm"
              variant="ghost"
              disabled={!allowed}
              title={
                allowed
                  ? `Add ${run} to the plan`
                  : `Your role, ${role.label}, does not have ${m.permission ?? "a permission that the UI knows"}`
              }
              onClick={() => {
                const def = step(run);
                if (!def) return;
                addStep(def);
                setAdded((n) => n + 1);
              }}
              className="border border-border"
            >
              <ListPlus className="h-4 w-4" />
              Add to plan
            </Button>
          </div>
        )}
      </div>

      {m.long && (
        <Text variant="muted" as="p" className="mt-2 max-w-[80ch] pl-7">
          {m.long}
        </Text>
      )}

      <div className="mt-3 space-y-3 pl-7">
        <div className="space-y-0.5 rounded-md border border-border/60 bg-surface px-3 py-2 font-mono text-xs">
          <p>
            <span className="inline-block w-12 text-text-muted">in</span>
            <span className="text-text">{m.input}</span>
          </p>
          <p>
            <span className="inline-block w-12 text-text-muted">out</span>
            <span className="text-text">{m.output}</span>
          </p>
          {m.path && (
            <p>
              <span className="inline-block w-12 text-text-muted">route</span>
              <span className="font-semibold text-text">
                {m.httpMethod}
              </span>{" "}
              <span className="text-text-muted">{m.path}</span>
            </p>
          )}
        </div>

        <KVList
          className="gap-y-1.5"
          pairs={[
            [
              "Permission",
              m.permission ? (
                <span className="flex flex-wrap items-center gap-2">
                  <span className="font-mono">{m.permission}</span>
                  {allowed ? (
                    <span className="inline-flex items-center gap-1 text-text-muted">
                      <Check className="h-3 w-3" />
                      your role, {role.label}, has it
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1 text-text-muted">
                      <Lock className="h-3 w-3" />
                      your role, {role.label}, does not have it
                    </span>
                  )}
                </span>
              ) : (
                <span className="text-text-muted">
                  none: only another behavior calls it
                </span>
              ),
            ],
            ["Platforms", <Platforms names={m.platforms} />],
            [
              "Disrupts",
              m.disruptive ? (
                <span className="text-status-pending">
                  {m.disruption || "The host or its service stops for a time."}
                </span>
              ) : null,
            ],
          ]}
        />

        {m.calls.length > 0 && (
          <div>
            <Text variant="label" as="p" className="mb-1">
              Chain
            </Text>
            <CallTree behavior={b.name} method={m.name} calls={m.calls} />
          </div>
        )}

        {m.examples.length > 0 && (
          <CollapsibleSection
            open={examplesOpen}
            onToggle={() => setExamplesOpen(!examplesOpen)}
            label={`Examples (${m.examples.length})`}
            className="-mx-3"
          >
            <div className="space-y-3 px-3 pb-2">
              {m.examples.map((ex) => (
                <div key={ex.doc}>
                  <Text as="p" size="sm" className="mb-1">
                    {ex.doc}
                  </Text>
                  <div className="grid gap-2 lg:grid-cols-2">
                    <div>
                      <Text variant="muted" as="p" className="mb-0.5">
                        input
                      </Text>
                      <CodeBlock variant="json">
                        {JSON.stringify(ex.input, null, 2)}
                      </CodeBlock>
                    </div>
                    {ex.output !== undefined && (
                      <div>
                        <Text variant="muted" as="p" className="mb-0.5">
                          output
                        </Text>
                        <CodeBlock variant="json" maxHeight="16rem">
                          {JSON.stringify(ex.output, null, 2)}
                        </CodeBlock>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </CollapsibleSection>
        )}
      </div>
    </Card>
  );
}
