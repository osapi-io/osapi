import { useCallback, useMemo, useRef, useState } from "react";
import { FileCode } from "lucide-react";
import { useVimNav } from "@/hooks/use-vim-nav";
import { useStepFocus } from "@/hooks/use-step-focus";
import { useCommands } from "@/hooks/use-commands";
import { useAuth } from "@/hooks/use-auth";
import { useAgents } from "@/hooks/use-agents";
import { useCatalog } from "@/hooks/use-catalog";
import { usePlan } from "@/hooks/use-plan";
import { ContentArea } from "@/components/layout/content-area";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { PageHeader } from "@/components/ui/page-header";
import { StepCard } from "@/components/domain/step-card";
import { StepForm } from "@/components/domain/step-form";
import { StepPicker } from "@/components/domain/step-picker";
import { PlanPreflight } from "@/components/domain/plan-preflight";
import { PlanYAML } from "@/components/domain/plan-yaml";
import { ResultCard } from "@/components/domain/result-card";
import { RunLinks } from "@/components/domain/run-links";
import { JobChain } from "@/components/domain/job-chain";
import { RunBar } from "@/components/domain/run-bar";
import { ObjectUploadForm } from "@/components/domain/object-upload-form";
import { ObjectDeleteForm } from "@/components/domain/object-delete-form";
import {
  deleteFileByName,
  getFileStale,
  getFiles,
  postFile,
} from "@/sdk/gen/file-management-api-file-operations/file-management-api-file-operations";
import { runMethod } from "@/lib/run-method";
import type { StepDef } from "@/lib/catalog-index";
import { permitted } from "@/lib/permissions";
import {
  planYAML,
  preflight,
  stepRunnable,
  stepStatus,
  type PlanStep,
} from "@/lib/plan";
import { traceIdOf } from "@/lib/trace";

/** Runs a call of the controller: a step of the object store. */
async function runController(step: PlanStep): Promise<{ data: unknown }> {
  switch (step.run) {
    case "objectstore.list":
      return getFiles();
    case "objectstore.upload":
      return postFile({
        name: step.with.name as string,
        content_type: ((step.with.content_type as string) || "raw") as
          "raw" | "template",
        file: step.with._file as globalThis.File,
      });
    case "objectstore.delete":
      return deleteFileByName(step.with.name as string);
    default:
      return getFileStale();
  }
}

/**
 * Plan is the composer: steps that run in order, each a method of a behavior
 * with its target and its input. A read starts on each agent; a write or an
 * action waits for a target. Run runs each step that can run, in order, and
 * keeps going after a step fails. A plan is not saved: it exports as YAML.
 */
export function Plan() {
  const { can, role } = useAuth();
  const { agents } = useAgents(30000);
  const { steps: defs, step: stepDef } = useCatalog();
  const {
    steps,
    addStep,
    removeStep,
    moveStep,
    setWith,
    setTarget,
    setOutcome,
    resetSteps,
    clearSteps,
  } = usePlan();
  const [running, setRunning] = useState(false);
  const [yamlOpen, setYamlOpen] = useState(false);
  const [focusedId, setFocusedId] = useState<string | null>(null);
  const stepRefs = useRef<Map<string, HTMLDivElement>>(new Map());
  const resultRefs = useRef<Map<string, HTMLDivElement>>(new Map());
  const focusStep = useStepFocus(stepRefs);

  const allowed = useCallback(
    (permission: string | undefined) => permitted(permission, can),
    [can],
  );
  const runnable = (s: PlanStep) => {
    const def = stepDef(s.run);
    return def !== undefined && stepRunnable(def, s);
  };
  const canRun = steps.some(runnable);
  const hasRun = steps.some(
    (s) => s.outcome === "done" || s.outcome === "error",
  );
  const ok = steps.filter((s) => s.outcome === "done").length;
  const failed = steps.filter((s) => s.outcome === "error").length;
  const summary = hasRun
    ? `${ok + failed} ran: ${ok} ok${failed ? `, ${failed} failed` : ""}`
    : `${steps.length} ${steps.length === 1 ? "step" : "steps"}`;

  const check = useMemo(
    () => preflight(steps, stepDef, (p) => allowed(p), agents),
    [steps, stepDef, allowed, agents],
  );

  const add = useCallback(
    (def: StepDef) => {
      const id = addStep(def);
      setFocusedId(id);
      focusStep(id);
    },
    [addStep, focusStep],
  );

  const scrollTo = useCallback(async (id: string) => {
    await new Promise((r) => setTimeout(r, 50));
    stepRefs.current
      .get(id)
      ?.scrollIntoView({ behavior: "smooth", block: "center" });
  }, []);

  const scrollToResult = useCallback(async (id: string) => {
    await new Promise((r) => setTimeout(r, 150));
    const el = resultRefs.current.get(id);
    if (!el) return;
    // The result sits a third of the way down, so the step above it shows.
    const y =
      el.getBoundingClientRect().top +
      window.scrollY -
      window.innerHeight * 0.3;
    window.scrollTo({ top: y, behavior: "smooth" });
  }, []);

  const handleRun = useCallback(async () => {
    setRunning(true);
    for (const step of steps) {
      const def = stepDef(step.run);
      if (!def || !stepRunnable(def, step)) continue;
      setFocusedId(step.id);
      await scrollTo(step.id);
      setOutcome(step.id, "running");
      try {
        const result = def.method
          ? await runMethod(def.method, step.target, step.with)
          : await runController(step);
        setOutcome(step.id, "done", result.data, undefined, traceIdOf(result));
      } catch (err) {
        setOutcome(
          step.id,
          "error",
          undefined,
          err instanceof Error ? err.message : "Unknown error",
          (err as { traceId?: string } | undefined)?.traceId,
        );
      }
      await scrollToResult(step.id);
    }
    setFocusedId(null);
    setRunning(false);
  }, [steps, stepDef, setOutcome, scrollTo, scrollToResult]);

  useCommands(
    [
      {
        id: "cmd:run",
        name: "run",
        description: "Run the plan",
        category: "actions",
        action: () => {
          if (canRun && !running) handleRun();
        },
      },
      {
        id: "cmd:clear",
        name: "clear",
        description: "Remove each step",
        category: "actions",
        action: () => clearSteps(),
      },
      {
        id: "cmd:yaml",
        name: "yaml",
        description: "Show the plan as YAML",
        category: "actions",
        action: () => setYamlOpen(true),
      },
      ...defs
        .filter((d) => allowed(d.permission))
        .map((d) => ({
          id: `step:${d.run}`,
          name: d.run.replace(".", " "),
          description: d.doc,
          category: "steps",
          action: () => add(d),
        })),
    ],
    [canRun, running, handleRun, clearSteps, defs, allowed, add],
  );

  useVimNav({
    items: steps,
    focusedId,
    setFocusedId,
    onDelete: removeStep,
    onExecute: handleRun,
    canExecute: canRun,
    executing: running,
    itemRefs: stepRefs,
  });

  const renderForm = (step: PlanStep, def: StepDef) => {
    if (def.method) {
      return (
        <StepForm
          stepId={step.id}
          method={def.method}
          data={step.with}
          onChange={(data) => setWith(step.id, data)}
        />
      );
    }
    if (step.run === "objectstore.upload") {
      return (
        <ObjectUploadForm
          data={step.with}
          onChange={(data) => setWith(step.id, data)}
        />
      );
    }
    if (step.run === "objectstore.delete") {
      return (
        <ObjectDeleteForm
          stepId={step.id}
          data={step.with}
          onChange={(data) => setWith(step.id, data)}
        />
      );
    }
    return null;
  };

  const items: React.ReactNode[] = [];
  steps.forEach((step, i) => {
    const def = stepDef(step.run);
    if (!def) return;
    items.push(
      <StepCard
        key={step.id}
        ref={(el) => {
          if (el) stepRefs.current.set(step.id, el);
          else stepRefs.current.delete(step.id);
        }}
        index={i + 1}
        def={def}
        status={stepStatus(def, step)}
        focused={focusedId === step.id}
        error={step.error}
        target={step.target}
        onTargetChange={(t) => setTarget(step.id, t)}
        onRemove={() => removeStep(step.id)}
        onMove={(by) => moveStep(step.id, by)}
        first={i === 0}
        last={i === steps.length - 1}
        onFocusCard={() => setFocusedId(step.id)}
      >
        {renderForm(step, def)}
      </StepCard>,
    );
    const jobId = (step.result as { job_id?: unknown } | undefined)?.job_id;
    if (step.result != null || step.traceId) {
      items.push(
        <RunLinks
          key={`${step.id}-links`}
          jobId={typeof jobId === "string" ? jobId : undefined}
          traceId={step.traceId}
          className="-mt-1 mb-1 px-1"
        />,
      );
    }
    if (step.result != null) {
      items.push(
        <div
          key={`${step.id}-result`}
          className="space-y-2"
          ref={(el) => {
            if (el) resultRefs.current.set(step.id, el);
            else resultRefs.current.delete(step.id);
          }}
        >
          <ResultCard
            run={step.run}
            method={def.method}
            title={def.title}
            result={step.result}
            onAction={async (action, name) => {
              if (action !== "objectstore.delete") return;
              try {
                await deleteFileByName(name);
              } catch {
                // The list does not change, which shows that it failed.
              }
            }}
          />
          {typeof jobId === "string" && <JobChain jobId={jobId} />}
        </div>,
      );
    }
  });

  return (
    <ContentArea>
      <PageHeader
        title="New plan"
        subtitle="Steps run in order. A read starts on each agent; a write or an action runs only on the hosts that you pick."
        actions={
          <Button
            variant="ghost"
            size="sm"
            className="border border-border"
            disabled={steps.length === 0}
            onClick={() => setYamlOpen(true)}
          >
            <FileCode className="h-4 w-4" />
            Export YAML
          </Button>
        }
      />

      <div className="grid grid-cols-12 gap-6">
        {/* The list stays in view as the plan grows, and scrolls on its own
            when it is taller than the window, so each method is reachable. */}
        <aside className="col-span-12 self-start md:sticky md:top-20 md:col-span-4 md:max-h-[calc(100vh-6rem)] md:overflow-y-auto md:pr-1">
          <StepPicker steps={defs} allowed={allowed} onAdd={add} />
        </aside>

        <main className="col-span-12 md:col-span-8">
          {steps.length === 0 ? (
            <EmptyState message="Add a step from the list. Open Behaviors to read what each method does first." />
          ) : (
            <>
              <PlanPreflight check={check} role={role.label} />
              <div className="space-y-3">{items}</div>
              <RunBar
                summary={summary}
                ready={canRun}
                running={running}
                hasRun={hasRun}
                onRun={handleRun}
                onReset={resetSteps}
                onClear={clearSteps}
              />
              <div className="h-24" />
            </>
          )}
        </main>
      </div>

      <PlanYAML
        open={yamlOpen}
        onClose={() => setYamlOpen(false)}
        yaml={planYAML(steps, stepDef)}
      />
    </ContentArea>
  );
}
