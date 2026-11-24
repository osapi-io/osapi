import { useMemo, useState } from "react";
import { ListBox } from "@/components/ui/list-box";
import { Link } from "react-router-dom";
import { ChevronDown, ChevronRight, Info, Lock, Plus } from "lucide-react";
import { SearchBox } from "@/components/ui/search-box";
import { Text } from "@/components/ui/text";
import { KindMark } from "@/components/domain/method-kind";
import { AREAS } from "@/lib/catalog.gen";
import { stepMatches, type StepDef } from "@/lib/catalog-index";

interface StepPickerProps {
  steps: StepDef[];
  allowed: (permission: string | undefined) => boolean;
  onAdd: (def: StepDef) => void;
}

/** The steps of one behavior, or of the controller's object store. */
interface Group {
  key: string;
  title: string;
  /** The behavior's page, when the steps are a behavior's. */
  behavior?: string;
  steps: StepDef[];
}

/**
 * StepPicker is the list that a plan's steps come from: each area, then
 * each behavior in it, then each method of the behavior. A behavior opens to
 * show its methods; a click on a method adds it to the plan. A filter opens
 * each behavior that matches. A method that the viewer's role cannot run is
 * locked.
 */
export function StepPicker({ steps, allowed, onAdd }: StepPickerProps) {
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState<Set<string>>(new Set());

  const areas = useMemo(() => {
    const shown = steps.filter((s) => stepMatches(s, query));
    return AREAS.map((a) => {
      const groups: Group[] = [];
      for (const s of shown.filter((x) => x.area === a.name)) {
        const key = s.behavior?.name ?? `${a.name}-controller`;
        let g = groups.find((x) => x.key === key);
        if (!g) {
          g = {
            key,
            title: s.title,
            behavior: s.behavior?.name,
            steps: [],
          };
          groups.push(g);
        }
        g.steps.push(s);
      }
      return { ...a, groups };
    }).filter((a) => a.groups.length > 0);
  }, [steps, query]);

  const toggle = (key: string) =>
    setOpen((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });

  return (
    <div className="space-y-3">
      <SearchBox
        value={query}
        onChange={setQuery}
        onClose={() => setQuery("")}
        placeholder="Find a method, such as dns update"
        autoFocus={false}
        inputClassName="min-w-0 flex-1"
      />
      {areas.length === 0 && (
        <Text variant="muted" as="p" className="px-1">
          No method matches "{query}".
        </Text>
      )}
      {areas.map((a) => (
        <div key={a.name}>
          <h3 className="mb-1 px-1 text-xs font-semibold text-text-muted">
            {a.label}
          </h3>
          <ListBox>
            {a.groups.map((g) => {
              const isOpen = query !== "" || open.has(g.key);
              return (
                <li
                  key={g.key}
                  className="border-b border-border last:border-0"
                >
                  <button
                    type="button"
                    onClick={() => toggle(g.key)}
                    aria-expanded={isOpen}
                    className="flex w-full items-center gap-2 px-3 py-2 text-left transition-colors hover:bg-white/[0.03]"
                  >
                    {isOpen ? (
                      <ChevronDown className="h-3.5 w-3.5 shrink-0 text-text-muted" />
                    ) : (
                      <ChevronRight className="h-3.5 w-3.5 shrink-0 text-text-muted" />
                    )}
                    <span className="flex-1 truncate text-sm font-medium text-white">
                      {g.title}
                    </span>
                    <span className="flex gap-0.5">
                      {g.steps.map((s) => (
                        <KindMark
                          key={s.run}
                          kind={s.kind}
                          locked={!allowed(s.permission)}
                          title={s.run}
                        />
                      ))}
                    </span>
                  </button>
                  {isOpen && (
                    <ul className="border-t border-border/60 bg-surface py-1">
                      {g.steps.map((s) => (
                        <MethodRow
                          key={s.run}
                          step={s}
                          allowed={allowed(s.permission)}
                          onAdd={() => onAdd(s)}
                        />
                      ))}
                      {g.behavior && (
                        <li className="px-3 pb-1 pt-0.5">
                          <Link
                            to={`/behaviors/${g.behavior}`}
                            className="inline-flex items-center gap-1 pl-5 text-xs text-text-muted hover:text-text"
                          >
                            <Info className="h-3 w-3" />
                            About {g.title}
                          </Link>
                        </li>
                      )}
                    </ul>
                  )}
                </li>
              );
            })}
          </ListBox>
        </div>
      ))}
    </div>
  );
}

function MethodRow({
  step: s,
  allowed,
  onAdd,
}: {
  step: StepDef;
  allowed: boolean;
  onAdd: () => void;
}) {
  return (
    <li>
      <button
        type="button"
        onClick={onAdd}
        disabled={!allowed}
        title={
          allowed
            ? `Add ${s.run}: ${s.doc}`
            : `Needs ${s.permission ?? "a permission that the UI knows"}`
        }
        className="group flex w-full items-center gap-2 py-1.5 pl-8 pr-3 text-left transition-colors hover:bg-white/[0.04] disabled:cursor-not-allowed disabled:opacity-40"
      >
        <KindMark kind={s.kind} />
        <span className="font-mono text-xs text-text">
          {s.label.toLowerCase()}
        </span>
        <span className="flex-1 truncate text-xs text-text-muted">{s.doc}</span>
        {allowed ? (
          <Plus className="h-3.5 w-3.5 shrink-0 text-text-muted/40 group-hover:text-text" />
        ) : (
          <Lock className="h-3 w-3 shrink-0 text-text-muted" />
        )}
      </button>
    </li>
  );
}
