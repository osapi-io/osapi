import { useState } from "react";
import {
  ChevronDown,
  ChevronRight,
  Code,
  Trash2,
  Copy,
  Check,
} from "lucide-react";
import { SearchBox } from "@/components/ui/search-box";
import type {
  CommandResultItem,
  CronMutationResult,
  DockerResponse,
  FileDeployResult,
  FileInfo,
  FileUploadResponse,
} from "@/sdk/gen/schemas";
import { JobDetail } from "@/components/domain/job-detail";
import { CodeBlock } from "@/components/ui/code-block";
import { KeyValue } from "@/components/ui/key-value";
import { StatusIcon } from "@/components/ui/status-icon";
import { Text } from "@/components/ui/text";
import { CollapsibleSection } from "@/components/ui/collapsible-section";
import { IconButton } from "@/components/ui/icon-button";
import { HostGroupHeader } from "@/components/domain/host-group-header";

interface ResultCardProps {
  type: string;
  result: unknown;
  onAction?: (action: string, name: string) => void;
}

// ---------------------------------------------------------------------------
// Shared
// ---------------------------------------------------------------------------

/** Derive status from a result that has the status field. */
function hostStatus(r: {
  status?: string;
  error?: string;
}): "ok" | "failed" | "skipped" {
  if (r.status === "skipped") return "skipped";
  if (r.status === "failed" || r.error) return "failed";
  return "ok";
}

function ChangedLabel({ changed }: { changed?: boolean }) {
  if (changed === undefined || changed === null) return null;
  return (
    <Text variant={changed ? "primary" : "label"}>
      {changed ? "changed" : "unchanged"}
    </Text>
  );
}

function RawJsonToggle({ data }: { data: unknown }) {
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  const json = JSON.stringify(data, null, 2);

  const handleCopy = async (e: React.MouseEvent) => {
    e.stopPropagation();
    await navigator.clipboard.writeText(json);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <CollapsibleSection
      open={open}
      onToggle={() => setOpen(!open)}
      icon={Code}
      label="JSON"
      rightContent={
        open ? (
          <span
            onClick={handleCopy}
            className="flex items-center gap-1 rounded px-1.5 py-0.5 hover:bg-white/5"
          >
            {copied ? (
              <Check className="h-3 w-3 text-primary" />
            ) : (
              <Copy className="h-3 w-3" />
            )}
            {copied ? "Copied" : "Copy"}
          </span>
        ) : undefined
      }
    >
      <CodeBlock variant="json">{json}</CodeBlock>
    </CollapsibleSection>
  );
}

function ResultShell({
  jobId,
  label,
  summary,
  children,
  data,
}: {
  jobId?: string;
  label: string;
  summary: React.ReactNode;
  children?: React.ReactNode;
  data: unknown;
}) {
  return (
    <div className="rounded-lg border border-border/60 bg-[#030303] shadow-[inset_0_1px_0_rgba(255,255,255,0.02)]">
      <div className="flex items-center gap-2 border-b border-border/40 px-4 py-2.5">
        {summary}
        <Text
          variant="muted"
          className="font-semibold uppercase tracking-wider"
        >
          {label}
        </Text>
        {jobId && (
          <div className="ml-auto">
            <JobDetail jobId={jobId} inline />
          </div>
        )}
      </div>
      {jobId && <JobDetail jobId={jobId} panel />}
      {children}
      <RawJsonToggle data={data} />
    </div>
  );
}

// ---------------------------------------------------------------------------
// Command
// ---------------------------------------------------------------------------

function CommandHostRow({ host }: { host: CommandResultItem }) {
  const [expanded, setExpanded] = useState(true);
  const status = hostStatus(host);
  const hasOutput = host.stdout || host.stderr || host.error;

  return (
    <div className="border-b border-border/50 last:border-0">
      <button
        onClick={() => hasOutput && setExpanded(!expanded)}
        className="flex w-full items-center gap-2 px-4 py-2 text-left"
      >
        <StatusIcon status={status} />
        <Text className="font-medium">{host.hostname}</Text>
        <div className="ml-auto flex items-center gap-2">
          {host.duration_ms !== undefined && (
            <Text variant="muted">{host.duration_ms}ms</Text>
          )}
          <Text
            variant={
              status === "ok"
                ? "mono-primary"
                : status === "skipped"
                  ? "muted"
                  : "error"
            }
            className={status === "ok" ? undefined : "font-mono"}
          >
            {status === "skipped"
              ? "skipped"
              : host.error
                ? "error"
                : `exit ${host.exit_code ?? 0}`}
          </Text>
          {hasOutput &&
            (expanded ? (
              <ChevronDown className="h-3 w-3 text-text-muted" />
            ) : (
              <ChevronRight className="h-3 w-3 text-text-muted" />
            ))}
        </div>
      </button>
      {expanded && hasOutput && (
        <div className="px-4 pb-2">
          {host.error && <CodeBlock variant="error">{host.error}</CodeBlock>}
          {host.stdout && <CodeBlock variant="stdout">{host.stdout}</CodeBlock>}
          {host.stderr && (
            <CodeBlock variant="stderr" className="mt-1">
              {host.stderr}
            </CodeBlock>
          )}
        </div>
      )}
    </div>
  );
}

function CommandResult({
  data,
}: {
  data: { job_id?: string; results: CommandResultItem[] };
}) {
  const allOk = data.results.every((r) => hostStatus(r) === "ok");
  return (
    <ResultShell
      jobId={data.job_id}
      label={`${data.results.length} host${data.results.length !== 1 ? "s" : ""}`}
      summary={
        <div
          className={`h-1.5 w-1.5 rounded-full ${allOk ? "bg-primary" : "bg-status-error"}`}
        />
      }
      data={data}
    >
      {data.results.map((host, i) => (
        <CommandHostRow key={i} host={host} />
      ))}
    </ResultShell>
  );
}

// ---------------------------------------------------------------------------
// Host-level results (cron, file-deploy, docker)
// ---------------------------------------------------------------------------

function HostResultRow({
  hostname,
  status = "ok",
  changed,
  error,
  extra,
}: {
  hostname?: string;
  status?: "ok" | "failed" | "skipped";
  changed?: boolean;
  error?: string;
  extra?: React.ReactNode;
}) {
  return (
    <div className="border-b border-border/50 px-4 py-2 last:border-0">
      <div className="flex items-center gap-2">
        <StatusIcon status={status} />
        <Text className="font-medium">{hostname ?? "unknown"}</Text>
        {status === "skipped" && (
          <Text variant="muted" className="font-mono">
            skipped
          </Text>
        )}
        {extra}
        <div className="ml-auto">
          <ChangedLabel changed={changed} />
        </div>
      </div>
      {error && status !== "skipped" && (
        <Text variant="error" as="p" className="mt-1 pl-5">
          {error}
        </Text>
      )}
    </div>
  );
}

function CollectionResult({
  jobId,
  label,
  results,
  data,
}: {
  jobId?: string;
  label: string;
  results: {
    hostname?: string;
    status?: "ok" | "failed" | "skipped";
    changed?: boolean;
    error?: string;
    extra?: React.ReactNode;
  }[];
  data: unknown;
}) {
  const allOk = results.every((r) => r.status !== "failed");
  return (
    <ResultShell
      jobId={jobId}
      label={`${results.length} host${results.length !== 1 ? "s" : ""} · ${label}`}
      summary={
        <div
          className={`h-1.5 w-1.5 rounded-full ${allOk ? "bg-primary" : "bg-status-error"}`}
        />
      }
      data={data}
    >
      {results.map((r, i) => (
        <HostResultRow key={i} {...r} />
      ))}
    </ResultShell>
  );
}

// ---------------------------------------------------------------------------
// File Upload (single result, no hosts)
// ---------------------------------------------------------------------------

function FileUploadResult({ data }: { data: FileUploadResponse }) {
  return (
    <ResultShell
      label="uploaded"
      summary={
        <div
          className={`h-1.5 w-1.5 rounded-full ${data.changed ? "bg-primary" : "bg-text-muted"}`}
        />
      }
      data={data}
    >
      <div className="flex items-center gap-4 px-4 py-2.5">
        <StatusIcon status="ok" />
        <Text className="font-medium">{data.name}</Text>
        <Text variant="muted">{formatSize(data.size)}</Text>
        <Text variant="muted">{data.content_type}</Text>
        <ChangedLabel changed={data.changed} />
        <Text variant="mono-muted" className="ml-auto">
          sha:{data.sha256.slice(0, 12)}
        </Text>
      </div>
    </ResultShell>
  );
}

// ---------------------------------------------------------------------------
// List renderers (expanded, with search)
// ---------------------------------------------------------------------------

function useListSearch() {
  const [search, setSearch] = useState("");
  return { search, setSearch };
}

function matches(
  term: string,
  ...fields: (string | number | boolean | undefined | null)[]
) {
  if (!term) return true;
  const t = term.toLowerCase();
  return fields.some((f) => f != null && String(f).toLowerCase().includes(t));
}

function ListSearchBar({
  search,
  setSearch,
  hasItems,
}: {
  search: string;
  setSearch: (v: string) => void;
  hasItems: boolean;
}) {
  if (!hasItems) return null;
  return (
    <div className="border-b border-border/40 px-4 py-2">
      <SearchBox
        value={search}
        onChange={setSearch}
        onClose={() => setSearch("")}
        placeholder="Filter..."
        autoFocus={false}
      />
    </div>
  );
}

// Cron list
function CronListResult({
  data,
  onAction,
}: {
  data: Record<string, unknown>;
  onAction?: (action: string, name: string) => void;
}) {
  const { search, setSearch } = useListSearch();
  type CronListEntry = {
    hostname?: string;
    status?: string;
    name?: string;
    schedule?: string;
    object?: string;
    source?: string;
    error?: string;
  };
  const entries = data.results as CronListEntry[];
  const byHost = new Map<string, CronListEntry[]>();
  for (const e of entries) {
    const h = e.hostname ?? "unknown";
    if (!byHost.has(h)) byHost.set(h, []);
    byHost.get(h)!.push(e);
  }
  const totalEntries = entries.filter((e) => e.name).length;
  return (
    <ResultShell
      jobId={data.job_id as string | undefined}
      label={`${byHost.size} host${byHost.size !== 1 ? "s" : ""} · cron list`}
      summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
      data={data}
    >
      <ListSearchBar
        search={search}
        setSearch={setSearch}
        hasItems={totalEntries > 0}
      />
      {[...byHost.entries()].map(([hostname, items]) => {
        const hostEntry = items[0];
        const isSkipped = hostEntry?.status === "skipped";
        const hostError = !isSkipped && items.find((e) => e.error && !e.name);
        const allCronEntries = isSkipped ? [] : items.filter((e) => e.name);
        const cronEntries = allCronEntries.filter((e) =>
          matches(search, e.name, e.schedule, e.object, e.source),
        );
        if (search && cronEntries.length === 0 && !isSkipped && !hostError)
          return null;
        return (
          <div
            key={hostname}
            className="border-b border-border/50 last:border-0"
          >
            <HostGroupHeader
              hostname={hostname}
              status={isSkipped ? "skipped" : hostError ? "failed" : "ok"}
              detail={
                isSkipped
                  ? "skipped"
                  : `${allCronEntries.length} entr${allCronEntries.length !== 1 ? "ies" : "y"}`
              }
            />
            {hostError && (
              <Text variant="error" as="p" className="px-4 pb-2">
                {(hostError as CronListEntry).error}
              </Text>
            )}
            {cronEntries.map((e, i) => (
              <div
                key={i}
                className="flex items-center gap-4 border-b border-border/30 px-4 py-2 pl-9 last:border-0"
              >
                <KeyValue label="name" value={e.name} variant="strong" />
                <KeyValue label="schedule" value={e.schedule} />
                {e.object && (
                  <KeyValue label="object" value={e.object} variant="accent" />
                )}
                {e.source && <KeyValue label="source" value={e.source} />}
                {onAction && e.name && (
                  <IconButton
                    icon={Trash2}
                    variant="danger"
                    onClick={() => onAction("cron-delete", e.name!)}
                    title={`Delete ${e.name}`}
                    className="ml-auto"
                  />
                )}
              </div>
            ))}
          </div>
        );
      })}
    </ResultShell>
  );
}

// Docker list
function DockerListResult({ data }: { data: Record<string, unknown> }) {
  const { search, setSearch } = useListSearch();
  type DockerHost = {
    hostname: string;
    status?: string;
    containers?: {
      id?: string;
      name?: string;
      image?: string;
      state?: string;
    }[];
    error?: string;
  };
  const hosts = data.results as DockerHost[];
  const totalContainers = hosts.reduce(
    (n, h) => n + (h.containers?.length ?? 0),
    0,
  );
  return (
    <ResultShell
      jobId={data.job_id as string | undefined}
      label={`${hosts.length} host${hosts.length !== 1 ? "s" : ""} · ${totalContainers} container${totalContainers !== 1 ? "s" : ""}`}
      summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
      data={data}
    >
      <ListSearchBar
        search={search}
        setSearch={setSearch}
        hasItems={totalContainers > 0}
      />
      {hosts.map((host) => {
        const allContainers = host.containers ?? [];
        const containers = allContainers.filter((c) =>
          matches(search, c.name, c.id, c.image, c.state),
        );
        if (
          search &&
          containers.length === 0 &&
          !host.error &&
          host.status !== "skipped"
        )
          return null;
        return (
          <div
            key={host.hostname}
            className="border-b border-border/50 last:border-0"
          >
            <HostGroupHeader
              hostname={host.hostname}
              status={
                host.status === "skipped"
                  ? "skipped"
                  : host.error
                    ? "failed"
                    : "ok"
              }
              detail={
                host.status === "skipped"
                  ? "skipped"
                  : `${allContainers.length} container${allContainers.length !== 1 ? "s" : ""}`
              }
            />
            {host.error && host.status !== "skipped" && (
              <Text variant="error" as="p" className="px-4 pb-2">
                {host.error}
              </Text>
            )}
            {host.status !== "skipped" &&
              containers.map((c, i) => (
                <div
                  key={i}
                  className="flex items-center gap-4 border-b border-border/30 px-4 py-2 pl-9 last:border-0"
                >
                  <KeyValue
                    label="name"
                    value={c.name || c.id?.slice(0, 12)}
                    variant="strong"
                  />
                  {c.image && <KeyValue label="image" value={c.image} />}
                  {c.state && (
                    <KeyValue
                      label="state"
                      value={c.state}
                      variant={c.state === "running" ? "accent" : "default"}
                      className="ml-auto"
                    />
                  )}
                </div>
              ))}
          </div>
        );
      })}
    </ResultShell>
  );
}

// Service list
function ServiceListResult({ data }: { data: Record<string, unknown> }) {
  const { search, setSearch } = useListSearch();
  type ServiceListEntry = {
    hostname: string;
    status?: string;
    services?: {
      name?: string;
      status?: string;
      enabled?: boolean;
      description?: string;
      pid?: number;
    }[];
    error?: string;
  };
  const hosts = data.results as ServiceListEntry[];
  const totalServices = hosts.reduce(
    (n, h) => n + (h.services?.length ?? 0),
    0,
  );
  return (
    <ResultShell
      jobId={data.job_id as string | undefined}
      label={`${hosts.length} host${hosts.length !== 1 ? "s" : ""} · ${totalServices} service${totalServices !== 1 ? "s" : ""}`}
      summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
      data={data}
    >
      <ListSearchBar
        search={search}
        setSearch={setSearch}
        hasItems={totalServices > 0}
      />
      {hosts.map((host) => {
        const allServices = host.services ?? [];
        const services = allServices.filter((s) =>
          matches(search, s.name, s.status, s.description),
        );
        if (
          search &&
          services.length === 0 &&
          !host.error &&
          host.status !== "skipped"
        )
          return null;
        return (
          <div
            key={host.hostname}
            className="border-b border-border/50 last:border-0"
          >
            <HostGroupHeader
              hostname={host.hostname}
              status={
                host.status === "skipped"
                  ? "skipped"
                  : host.error
                    ? "failed"
                    : "ok"
              }
              detail={
                host.status === "skipped"
                  ? "skipped"
                  : `${allServices.length} service${allServices.length !== 1 ? "s" : ""}`
              }
            />
            {host.error && host.status !== "skipped" && (
              <Text variant="error" as="p" className="px-4 pb-2">
                {host.error}
              </Text>
            )}
            {host.status !== "skipped" &&
              services.map((s, i) => (
                <div
                  key={i}
                  className="flex items-center gap-4 border-b border-border/30 px-4 py-2 pl-9 last:border-0"
                >
                  <KeyValue label="name" value={s.name} variant="strong" />
                  {s.status && (
                    <KeyValue
                      label="status"
                      value={s.status}
                      variant={s.status === "active" ? "accent" : "default"}
                    />
                  )}
                  {s.enabled !== undefined && (
                    <KeyValue
                      label="enabled"
                      value={s.enabled ? "yes" : "no"}
                    />
                  )}
                  {s.description && (
                    <Text variant="muted" className="truncate max-w-[200px]">
                      {s.description}
                    </Text>
                  )}
                </div>
              ))}
          </div>
        );
      })}
    </ResultShell>
  );
}

// Package list / get / check-updates
function PackageListResult({
  data,
  label,
}: {
  data: Record<string, unknown>;
  label: string;
}) {
  const { search, setSearch } = useListSearch();
  type PkgHost = {
    hostname: string;
    status?: string;
    packages?: {
      name?: string;
      version?: string;
      description?: string;
      size?: number;
    }[];
    error?: string;
  };
  const hosts = data.results as PkgHost[];
  const totalPackages = hosts.reduce(
    (n, h) => n + (h.packages?.length ?? 0),
    0,
  );
  return (
    <ResultShell
      jobId={data.job_id as string | undefined}
      label={`${hosts.length} host${hosts.length !== 1 ? "s" : ""} · ${label}`}
      summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
      data={data}
    >
      <ListSearchBar
        search={search}
        setSearch={setSearch}
        hasItems={totalPackages > 0}
      />
      {hosts.map((host) => {
        const allPackages = host.packages ?? [];
        const packages = allPackages.filter((p) =>
          matches(search, p.name, p.version, p.description),
        );
        if (
          search &&
          packages.length === 0 &&
          !host.error &&
          host.status !== "skipped"
        )
          return null;
        return (
          <div
            key={host.hostname}
            className="border-b border-border/50 last:border-0"
          >
            <HostGroupHeader
              hostname={host.hostname}
              status={
                host.status === "skipped"
                  ? "skipped"
                  : host.error
                    ? "failed"
                    : "ok"
              }
              detail={
                host.status === "skipped"
                  ? "skipped"
                  : `${allPackages.length} package${allPackages.length !== 1 ? "s" : ""}`
              }
            />
            {host.error && host.status !== "skipped" && (
              <Text variant="error" as="p" className="px-4 pb-2">
                {host.error}
              </Text>
            )}
            {host.status !== "skipped" &&
              packages.map((p, i) => (
                <div
                  key={i}
                  className="flex items-center gap-4 border-b border-border/30 px-4 py-2 pl-9 last:border-0"
                >
                  <KeyValue label="name" value={p.name} variant="strong" />
                  {p.version && <KeyValue label="version" value={p.version} />}
                  {p.description && (
                    <Text
                      variant="muted"
                      className="ml-auto truncate max-w-[200px]"
                    >
                      {p.description}
                    </Text>
                  )}
                </div>
              ))}
          </div>
        );
      })}
    </ResultShell>
  );
}

// Sysctl list — flat results grouped by hostname
function SysctlListResult({ data }: { data: Record<string, unknown> }) {
  const { search, setSearch } = useListSearch();
  type SysctlEntry = {
    hostname: string;
    status?: string;
    key?: string;
    value?: string;
    error?: string;
  };
  const entries = data.results as SysctlEntry[];
  const byHost = new Map<string, SysctlEntry[]>();
  for (const e of entries) {
    const h = e.hostname ?? "unknown";
    if (!byHost.has(h)) byHost.set(h, []);
    byHost.get(h)!.push(e);
  }
  const totalEntries = entries.filter((e) => e.key).length;
  return (
    <ResultShell
      jobId={data.job_id as string | undefined}
      label={`${byHost.size} host${byHost.size !== 1 ? "s" : ""} · sysctl list`}
      summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
      data={data}
    >
      <ListSearchBar
        search={search}
        setSearch={setSearch}
        hasItems={totalEntries > 0}
      />
      {[...byHost.entries()].map(([hostname, items]) => {
        const hostEntry = items[0];
        const isSkipped = hostEntry?.status === "skipped";
        const hostError = !isSkipped && items.find((e) => e.error && !e.key);
        const allKvEntries = isSkipped ? [] : items.filter((e) => e.key);
        const kvEntries = allKvEntries.filter((e) =>
          matches(search, e.key, e.value),
        );
        if (search && kvEntries.length === 0 && !isSkipped && !hostError)
          return null;
        return (
          <div
            key={hostname}
            className="border-b border-border/50 last:border-0"
          >
            <HostGroupHeader
              hostname={hostname}
              status={isSkipped ? "skipped" : hostError ? "failed" : "ok"}
              detail={
                isSkipped
                  ? "skipped"
                  : `${allKvEntries.length} entr${allKvEntries.length !== 1 ? "ies" : "y"}`
              }
            />
            {hostError && (
              <Text variant="error" as="p" className="px-4 pb-2">
                {(hostError as SysctlEntry).error}
              </Text>
            )}
            {kvEntries.map((e, i) => (
              <div
                key={i}
                className="flex items-center gap-4 border-b border-border/30 px-4 py-2 pl-9 last:border-0"
              >
                <KeyValue label="key" value={e.key} variant="strong" />
                <KeyValue label="value" value={e.value} />
              </div>
            ))}
          </div>
        );
      })}
    </ResultShell>
  );
}

// Interface list
function InterfaceListResult({ data }: { data: Record<string, unknown> }) {
  const { search, setSearch } = useListSearch();
  type InterfaceHost = {
    hostname: string;
    status?: string;
    interfaces?: {
      name?: string;
      ipv4?: string;
      ipv6?: string;
      mac?: string;
      state?: string;
      dhcp4?: boolean;
      mtu?: number;
      primary?: boolean;
    }[];
    error?: string;
  };
  const hosts = data.results as InterfaceHost[];
  const totalInterfaces = hosts.reduce(
    (n, h) => n + (h.interfaces?.length ?? 0),
    0,
  );
  return (
    <ResultShell
      jobId={data.job_id as string | undefined}
      label={`${hosts.length} host${hosts.length !== 1 ? "s" : ""} · interface list`}
      summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
      data={data}
    >
      <ListSearchBar
        search={search}
        setSearch={setSearch}
        hasItems={totalInterfaces > 0}
      />
      {hosts.map((host) => {
        const allInterfaces = host.interfaces ?? [];
        const interfaces = allInterfaces.filter((iface) =>
          matches(
            search,
            iface.name,
            iface.ipv4,
            iface.ipv6,
            iface.mac,
            iface.state,
          ),
        );
        if (
          search &&
          interfaces.length === 0 &&
          !host.error &&
          host.status !== "skipped"
        )
          return null;
        return (
          <div
            key={host.hostname}
            className="border-b border-border/50 last:border-0"
          >
            <HostGroupHeader
              hostname={host.hostname}
              status={
                host.status === "skipped"
                  ? "skipped"
                  : host.error
                    ? "failed"
                    : "ok"
              }
              detail={
                host.status === "skipped"
                  ? "skipped"
                  : `${allInterfaces.length} interface${allInterfaces.length !== 1 ? "s" : ""}`
              }
            />
            {host.error && host.status !== "skipped" && (
              <Text variant="error" as="p" className="px-4 pb-2">
                {host.error}
              </Text>
            )}
            {host.status !== "skipped" &&
              interfaces.map((iface, i) => (
                <div
                  key={i}
                  className="flex items-center gap-4 border-b border-border/30 px-4 py-2 pl-9 last:border-0"
                >
                  <KeyValue label="name" value={iface.name} variant="strong" />
                  {iface.ipv4 && <KeyValue label="ipv4" value={iface.ipv4} />}
                  {iface.state && (
                    <KeyValue
                      label="state"
                      value={iface.state}
                      variant={iface.state === "up" ? "accent" : "default"}
                    />
                  )}
                  {iface.mac && (
                    <KeyValue label="mac" value={iface.mac} variant="mono" />
                  )}
                </div>
              ))}
          </div>
        );
      })}
    </ResultShell>
  );
}

// Route list
function RouteListResult({ data }: { data: Record<string, unknown> }) {
  const { search, setSearch } = useListSearch();
  type RouteHost = {
    hostname: string;
    status?: string;
    routes?: {
      destination?: string;
      gateway?: string;
      interface?: string;
      metric?: number;
      scope?: string;
    }[];
    error?: string;
  };
  const hosts = data.results as RouteHost[];
  const totalRoutes = hosts.reduce((n, h) => n + (h.routes?.length ?? 0), 0);
  return (
    <ResultShell
      jobId={data.job_id as string | undefined}
      label={`${hosts.length} host${hosts.length !== 1 ? "s" : ""} · route list`}
      summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
      data={data}
    >
      <ListSearchBar
        search={search}
        setSearch={setSearch}
        hasItems={totalRoutes > 0}
      />
      {hosts.map((host) => {
        const allRoutes = host.routes ?? [];
        const routes = allRoutes.filter((r) =>
          matches(search, r.destination, r.gateway, r.interface, r.scope),
        );
        if (
          search &&
          routes.length === 0 &&
          !host.error &&
          host.status !== "skipped"
        )
          return null;
        return (
          <div
            key={host.hostname}
            className="border-b border-border/50 last:border-0"
          >
            <HostGroupHeader
              hostname={host.hostname}
              status={
                host.status === "skipped"
                  ? "skipped"
                  : host.error
                    ? "failed"
                    : "ok"
              }
              detail={
                host.status === "skipped"
                  ? "skipped"
                  : `${allRoutes.length} route${allRoutes.length !== 1 ? "s" : ""}`
              }
            />
            {host.error && host.status !== "skipped" && (
              <Text variant="error" as="p" className="px-4 pb-2">
                {host.error}
              </Text>
            )}
            {host.status !== "skipped" &&
              routes.map((r, i) => (
                <div
                  key={i}
                  className="flex items-center gap-4 border-b border-border/30 px-4 py-2 pl-9 last:border-0"
                >
                  <KeyValue
                    label="dest"
                    value={r.destination}
                    variant="strong"
                  />
                  {r.gateway && <KeyValue label="via" value={r.gateway} />}
                  {r.interface && <KeyValue label="dev" value={r.interface} />}
                  {r.metric != null && (
                    <KeyValue label="metric" value={String(r.metric)} />
                  )}
                </div>
              ))}
          </div>
        );
      })}
    </ResultShell>
  );
}

// User list
function UserListResult({ data }: { data: Record<string, unknown> }) {
  const { search, setSearch } = useListSearch();
  type UserHost = {
    hostname: string;
    status?: string;
    users?: {
      name?: string;
      uid?: number;
      gid?: number;
      home?: string;
      shell?: string;
      groups?: string[];
      locked?: boolean;
    }[];
    error?: string;
  };
  const hosts = data.results as UserHost[];
  const totalUsers = hosts.reduce((n, h) => n + (h.users?.length ?? 0), 0);
  return (
    <ResultShell
      jobId={data.job_id as string | undefined}
      label={`${hosts.length} host${hosts.length !== 1 ? "s" : ""} · user list`}
      summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
      data={data}
    >
      <ListSearchBar
        search={search}
        setSearch={setSearch}
        hasItems={totalUsers > 0}
      />
      {hosts.map((host) => {
        const allUsers = host.users ?? [];
        const users = allUsers.filter((u) =>
          matches(search, u.name, u.uid, u.shell, u.home),
        );
        if (
          search &&
          users.length === 0 &&
          !host.error &&
          host.status !== "skipped"
        )
          return null;
        return (
          <div
            key={host.hostname}
            className="border-b border-border/50 last:border-0"
          >
            <HostGroupHeader
              hostname={host.hostname}
              status={
                host.status === "skipped"
                  ? "skipped"
                  : host.error
                    ? "failed"
                    : "ok"
              }
              detail={
                host.status === "skipped"
                  ? "skipped"
                  : `${allUsers.length} user${allUsers.length !== 1 ? "s" : ""}`
              }
            />
            {host.error && host.status !== "skipped" && (
              <Text variant="error" as="p" className="px-4 pb-2">
                {host.error}
              </Text>
            )}
            {host.status !== "skipped" &&
              users.map((u, i) => (
                <div
                  key={i}
                  className="flex items-center gap-4 border-b border-border/30 px-4 py-2 pl-9 last:border-0"
                >
                  <KeyValue label="name" value={u.name} variant="strong" />
                  {u.uid != null && (
                    <KeyValue label="uid" value={String(u.uid)} />
                  )}
                  {u.shell && <KeyValue label="shell" value={u.shell} />}
                  {u.home && <KeyValue label="home" value={u.home} />}
                </div>
              ))}
          </div>
        );
      })}
    </ResultShell>
  );
}

// Group list
function GroupListResult({ data }: { data: Record<string, unknown> }) {
  const { search, setSearch } = useListSearch();
  type GroupHost = {
    hostname: string;
    status?: string;
    groups?: { name?: string; gid?: number; members?: string[] }[];
    error?: string;
  };
  const hosts = data.results as GroupHost[];
  const totalGroups = hosts.reduce((n, h) => n + (h.groups?.length ?? 0), 0);
  return (
    <ResultShell
      jobId={data.job_id as string | undefined}
      label={`${hosts.length} host${hosts.length !== 1 ? "s" : ""} · group list`}
      summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
      data={data}
    >
      <ListSearchBar
        search={search}
        setSearch={setSearch}
        hasItems={totalGroups > 0}
      />
      {hosts.map((host) => {
        const allGroups = host.groups ?? [];
        const groups = allGroups.filter((g) =>
          matches(search, g.name, g.gid, g.members?.join(" ")),
        );
        if (
          search &&
          groups.length === 0 &&
          !host.error &&
          host.status !== "skipped"
        )
          return null;
        return (
          <div
            key={host.hostname}
            className="border-b border-border/50 last:border-0"
          >
            <HostGroupHeader
              hostname={host.hostname}
              status={
                host.status === "skipped"
                  ? "skipped"
                  : host.error
                    ? "failed"
                    : "ok"
              }
              detail={
                host.status === "skipped"
                  ? "skipped"
                  : `${allGroups.length} group${allGroups.length !== 1 ? "s" : ""}`
              }
            />
            {host.error && host.status !== "skipped" && (
              <Text variant="error" as="p" className="px-4 pb-2">
                {host.error}
              </Text>
            )}
            {host.status !== "skipped" &&
              groups.map((g, i) => (
                <div
                  key={i}
                  className="flex items-center gap-4 border-b border-border/30 px-4 py-2 pl-9 last:border-0"
                >
                  <KeyValue label="name" value={g.name} variant="strong" />
                  {g.gid != null && (
                    <KeyValue label="gid" value={String(g.gid)} />
                  )}
                  <KeyValue
                    label="members"
                    value={String(g.members?.length ?? 0)}
                  />
                </div>
              ))}
          </div>
        );
      })}
    </ResultShell>
  );
}

// Certificate list
function CertificateListResult({ data }: { data: Record<string, unknown> }) {
  const { search, setSearch } = useListSearch();
  type CertHost = {
    hostname: string;
    status?: string;
    certificates?: { name?: string; source?: string; object?: string }[];
    error?: string;
  };
  const hosts = data.results as CertHost[];
  const totalCerts = hosts.reduce(
    (n, h) => n + (h.certificates?.length ?? 0),
    0,
  );
  return (
    <ResultShell
      jobId={data.job_id as string | undefined}
      label={`${hosts.length} host${hosts.length !== 1 ? "s" : ""} · certificate list`}
      summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
      data={data}
    >
      <ListSearchBar
        search={search}
        setSearch={setSearch}
        hasItems={totalCerts > 0}
      />
      {hosts.map((host) => {
        const allCerts = host.certificates ?? [];
        const certs = allCerts.filter((c) =>
          matches(search, c.name, c.source, c.object),
        );
        if (
          search &&
          certs.length === 0 &&
          !host.error &&
          host.status !== "skipped"
        )
          return null;
        return (
          <div
            key={host.hostname}
            className="border-b border-border/50 last:border-0"
          >
            <HostGroupHeader
              hostname={host.hostname}
              status={
                host.status === "skipped"
                  ? "skipped"
                  : host.error
                    ? "failed"
                    : "ok"
              }
              detail={
                host.status === "skipped"
                  ? "skipped"
                  : `${allCerts.length} cert${allCerts.length !== 1 ? "s" : ""}`
              }
            />
            {host.error && host.status !== "skipped" && (
              <Text variant="error" as="p" className="px-4 pb-2">
                {host.error}
              </Text>
            )}
            {host.status !== "skipped" &&
              certs.map((c, i) => (
                <div
                  key={i}
                  className="flex items-center gap-4 border-b border-border/30 px-4 py-2 pl-9 last:border-0"
                >
                  <KeyValue label="name" value={c.name} variant="strong" />
                  {c.source && <KeyValue label="source" value={c.source} />}
                </div>
              ))}
          </div>
        );
      })}
    </ResultShell>
  );
}

// Process list
function ProcessListResult({ data }: { data: Record<string, unknown> }) {
  const { search, setSearch } = useListSearch();
  type ProcessHost = {
    hostname: string;
    status?: string;
    processes?: {
      pid?: number;
      name?: string;
      user?: string;
      state?: string;
      cpu_percent?: number;
      mem_percent?: number;
      command?: string;
    }[];
    error?: string;
  };
  const hosts = data.results as ProcessHost[];
  const totalProcesses = hosts.reduce(
    (n, h) => n + (h.processes?.length ?? 0),
    0,
  );
  return (
    <ResultShell
      jobId={data.job_id as string | undefined}
      label={`${hosts.length} host${hosts.length !== 1 ? "s" : ""} · process list`}
      summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
      data={data}
    >
      <ListSearchBar
        search={search}
        setSearch={setSearch}
        hasItems={totalProcesses > 0}
      />
      {hosts.map((host) => {
        const allProcesses = host.processes ?? [];
        const processes = allProcesses.filter((p) =>
          matches(search, p.pid, p.name, p.user, p.state, p.command),
        );
        if (
          search &&
          processes.length === 0 &&
          !host.error &&
          host.status !== "skipped"
        )
          return null;
        return (
          <div
            key={host.hostname}
            className="border-b border-border/50 last:border-0"
          >
            <HostGroupHeader
              hostname={host.hostname}
              status={
                host.status === "skipped"
                  ? "skipped"
                  : host.error
                    ? "failed"
                    : "ok"
              }
              detail={
                host.status === "skipped"
                  ? "skipped"
                  : `${allProcesses.length} process${allProcesses.length !== 1 ? "es" : ""}`
              }
            />
            {host.error && host.status !== "skipped" && (
              <Text variant="error" as="p" className="px-4 pb-2">
                {host.error}
              </Text>
            )}
            {host.status !== "skipped" &&
              processes.map((p, i) => (
                <div
                  key={i}
                  className="flex items-center gap-4 border-b border-border/30 px-4 py-2 pl-9 last:border-0"
                >
                  {p.pid != null && (
                    <KeyValue label="pid" value={String(p.pid)} />
                  )}
                  <KeyValue label="name" value={p.name} variant="strong" />
                  {p.user && <KeyValue label="user" value={p.user} />}
                  {p.state && <KeyValue label="state" value={p.state} />}
                  {p.cpu_percent != null && (
                    <KeyValue
                      label="cpu"
                      value={`${p.cpu_percent.toFixed(1)}%`}
                    />
                  )}
                  {p.mem_percent != null && (
                    <KeyValue
                      label="mem"
                      value={`${p.mem_percent.toFixed(1)}%`}
                    />
                  )}
                </div>
              ))}
          </div>
        );
      })}
    </ResultShell>
  );
}

// Log sources list
function LogSourcesResult({ data }: { data: Record<string, unknown> }) {
  const { search, setSearch } = useListSearch();
  type LogSourceHost = {
    hostname: string;
    status?: string;
    sources?: string[];
    error?: string;
  };
  const hosts = data.results as LogSourceHost[];
  const totalSources = hosts.reduce((n, h) => n + (h.sources?.length ?? 0), 0);
  return (
    <ResultShell
      jobId={data.job_id as string | undefined}
      label={`${hosts.length} host${hosts.length !== 1 ? "s" : ""} · log sources`}
      summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
      data={data}
    >
      <ListSearchBar
        search={search}
        setSearch={setSearch}
        hasItems={totalSources > 0}
      />
      {hosts.map((host) => {
        const allSources = host.sources ?? [];
        const sources = allSources.filter((s) => matches(search, s));
        if (
          search &&
          sources.length === 0 &&
          !host.error &&
          host.status !== "skipped"
        )
          return null;
        return (
          <div
            key={host.hostname}
            className="border-b border-border/50 last:border-0"
          >
            <HostGroupHeader
              hostname={host.hostname}
              status={
                host.status === "skipped"
                  ? "skipped"
                  : host.error
                    ? "failed"
                    : "ok"
              }
              detail={
                host.status === "skipped"
                  ? "skipped"
                  : `${allSources.length} source${allSources.length !== 1 ? "s" : ""}`
              }
            />
            {host.error && host.status !== "skipped" && (
              <Text variant="error" as="p" className="px-4 pb-2">
                {host.error}
              </Text>
            )}
            {host.status !== "skipped" &&
              sources.map((s, i) => (
                <div
                  key={i}
                  className="border-b border-border/30 px-4 py-2 pl-9 last:border-0"
                >
                  <Text variant="mono-muted">{s}</Text>
                </div>
              ))}
          </div>
        );
      })}
    </ResultShell>
  );
}

// ---------------------------------------------------------------------------
// Router
// ---------------------------------------------------------------------------

export function ResultCard({ type, result, onAction }: ResultCardProps) {
  if (!result) return null;
  const data = result as Record<string, unknown>;

  if (
    (type === "command" ||
      type === "command-shell" ||
      type === "docker-exec") &&
    Array.isArray(data.results)
  ) {
    return (
      <CommandResult
        data={
          data as unknown as { job_id?: string; results: CommandResultItem[] }
        }
      />
    );
  }

  // File list
  if (type === "file-list" && "files" in data) {
    const files = (data as { files: FileInfo[] }).files;
    return (
      <ResultShell
        label={`${files.length} file${files.length !== 1 ? "s" : ""}`}
        summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
        data={data}
      >
        {files.map((f, i) => (
          <div
            key={i}
            className="flex items-center gap-4 border-b border-border/50 px-4 py-2 last:border-0"
          >
            <StatusIcon status="ok" />
            <KeyValue label="name" value={f.name} variant="strong" />
            <KeyValue label="size" value={formatSize(f.size)} />
            <KeyValue label="type" value={f.content_type} />
            {f.source && (
              <KeyValue label="source" value={f.source} variant="accent" />
            )}
            <Text variant="mono-muted">sha:{f.sha256.slice(0, 12)}</Text>
            {onAction && (
              <IconButton
                icon={Trash2}
                variant="danger"
                onClick={() => onAction("file-delete", f.name)}
                title={`Delete ${f.name}`}
                className="ml-auto"
              />
            )}
          </div>
        ))}
      </ResultShell>
    );
  }

  if (type === "file-upload" && "name" in data && "sha256" in data) {
    return <FileUploadResult data={data as unknown as FileUploadResponse} />;
  }

  // Cron create/delete — mutation results
  if (
    (type === "cron-create" || type === "cron-delete") &&
    Array.isArray(data.results)
  ) {
    const results = data.results as CronMutationResult[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label={type === "cron-create" ? "cron create" : "cron delete"}
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          changed: r.changed,
          error: r.error,
          extra: r.name ? <Text variant="muted">{r.name}</Text> : undefined,
        }))}
        data={data}
      />
    );
  }

  // Cron list — expanded with search
  if (type === "cron-list" && Array.isArray(data.results)) {
    return <CronListResult data={data} onAction={onAction} />;
  }

  // Docker create/start/stop/delete/pull/rm-image — action results
  if (
    (type === "docker-create" ||
      type === "docker-start" ||
      type === "docker-stop" ||
      type === "docker-delete" ||
      type === "docker-pull" ||
      type === "docker-rm-image") &&
    Array.isArray(data.results)
  ) {
    const results = data.results as DockerResponse[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label={type.replace("docker-", "")}
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          changed: r.changed,
          error: r.error,
          extra: (
            <span className="flex items-center gap-1.5">
              {r.name && <Text>{r.name}</Text>}
              {r.image && <Text variant="muted">{r.image}</Text>}
              {r.state && (
                <Text variant={r.state === "running" ? "primary" : "muted"}>
                  {r.state}
                </Text>
              )}
            </span>
          ),
        }))}
        data={data}
      />
    );
  }

  // Docker list — expanded with search
  if (type === "docker-list" && Array.isArray(data.results)) {
    return <DockerListResult data={data} />;
  }

  // File deploy/undeploy/status
  if (
    (type === "file-deploy" ||
      type === "file-undeploy" ||
      type === "file-status") &&
    Array.isArray(data.results)
  ) {
    const results = data.results as FileDeployResult[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label={type.replace("file-", "")}
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          changed: r.changed,
          error: r.error,
        }))}
        data={data}
      />
    );
  }

  // DNS list — show servers and search domains per host
  if (type === "dns-list" && Array.isArray(data.results)) {
    const results = data.results as {
      hostname: string;
      status?: string;
      servers?: string[];
      search_domains?: string[];
      error?: string;
    }[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label="dns"
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          error: r.error,
          extra: (
            <div className="space-y-0.5 text-xs text-text-muted">
              {r.servers && <p>servers: {r.servers.join(", ")}</p>}
              {r.search_domains && <p>search: {r.search_domains.join(", ")}</p>}
            </div>
          ),
        }))}
        data={data}
      />
    );
  }

  // DNS update — host results with status/changed
  if (type === "dns-update" && Array.isArray(data.results)) {
    const results = data.results as {
      hostname: string;
      status?: string;
      changed?: boolean;
      error?: string;
    }[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label="dns update"
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          changed: r.changed,
          error: r.error,
        }))}
        data={data}
      />
    );
  }

  // Interface list — expanded with search
  if (type === "interface-list" && Array.isArray(data.results)) {
    return <InterfaceListResult data={data} />;
  }

  // Route list — expanded with search
  if (type === "route-list" && Array.isArray(data.results)) {
    return <RouteListResult data={data} />;
  }

  // DNS delete / interface CRUD / route CRUD — generic collection results
  if (
    [
      "dns-delete",
      "interface-get",
      "interface-create",
      "interface-update",
      "interface-delete",
      "route-get",
      "route-create",
      "route-update",
      "route-delete",
    ].includes(type) &&
    Array.isArray(data.results)
  ) {
    const label = type.replace("-", " ");
    const results = data.results as {
      hostname?: string;
      status?: string;
      changed?: boolean;
      error?: string;
    }[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label={label}
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          changed: r.changed,
          error: r.error,
        }))}
        data={data}
      />
    );
  }

  // Ping — show RTT stats per host
  if (type === "ping" && Array.isArray(data.results)) {
    const results = data.results as {
      hostname: string;
      status?: string;
      packets_sent?: number;
      packets_received?: number;
      packet_loss?: number;
      min_rtt?: string;
      avg_rtt?: string;
      max_rtt?: string;
      error?: string;
    }[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label="ping"
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          error: r.error,
          extra: !r.error ? (
            <div className="flex gap-3 text-xs text-text-muted">
              <span>
                {r.packets_received}/{r.packets_sent} recv
              </span>
              {r.packet_loss != null && <span>{r.packet_loss}% loss</span>}
              {r.avg_rtt && <span>avg {r.avg_rtt}</span>}
              {r.min_rtt && r.max_rtt && (
                <span>
                  {r.min_rtt} – {r.max_rtt}
                </span>
              )}
            </div>
          ) : undefined,
        }))}
        data={data}
      />
    );
  }

  // Service list — expanded with search
  if (type === "service-list" && Array.isArray(data.results)) {
    return <ServiceListResult data={data} />;
  }

  // Service get — show per-host service details
  if (type === "service-get" && Array.isArray(data.results)) {
    type ServiceGetEntry = {
      hostname: string;
      status?: string;
      service?: {
        name?: string;
        status?: string;
        enabled?: boolean;
        description?: string;
        pid?: number;
      };
      error?: string;
    };
    const entries = data.results as ServiceGetEntry[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label="service get"
        results={entries.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          error: r.error,
          extra: r.service ? (
            <span className="flex items-center gap-1.5">
              {r.service.name && <Text>{r.service.name}</Text>}
              {r.service.status && (
                <Text
                  variant={r.service.status === "active" ? "primary" : "muted"}
                >
                  {r.service.status}
                </Text>
              )}
              {r.service.enabled !== undefined && (
                <Text variant="muted">
                  {r.service.enabled ? "enabled" : "disabled"}
                </Text>
              )}
            </span>
          ) : undefined,
        }))}
        data={data}
      />
    );
  }

  // Service mutations (create/update/delete/start/stop/restart/enable/disable)
  if (
    (type === "service-create" ||
      type === "service-update" ||
      type === "service-delete" ||
      type === "service-start" ||
      type === "service-stop" ||
      type === "service-restart" ||
      type === "service-enable" ||
      type === "service-disable") &&
    Array.isArray(data.results)
  ) {
    type ServiceMutEntry = {
      hostname: string;
      status?: string;
      name?: string;
      changed?: boolean;
      error?: string;
    };
    const results = data.results as ServiceMutEntry[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label={type.replace("service-", "")}
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          changed: r.changed,
          error: r.error,
          extra: r.name ? <Text variant="muted">{r.name}</Text> : undefined,
        }))}
        data={data}
      />
    );
  }

  // Cron get — per-host single entry
  if (type === "cron-get" && Array.isArray(data.results)) {
    type CronGetEntry = {
      hostname?: string;
      status?: string;
      name?: string;
      schedule?: string;
      object?: string;
      error?: string;
    };
    const entries = data.results as CronGetEntry[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label="cron get"
        results={entries.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          error: r.error,
          extra: r.name ? (
            <span className="flex items-center gap-1.5">
              <Text>{r.name}</Text>
              {r.schedule && <Text variant="muted">{r.schedule}</Text>}
            </span>
          ) : undefined,
        }))}
        data={data}
      />
    );
  }

  // Cron update — mutation results
  if (type === "cron-update" && Array.isArray(data.results)) {
    type CronUpdateEntry = {
      hostname?: string;
      status?: string;
      name?: string;
      changed?: boolean;
      error?: string;
    };
    const results = data.results as CronUpdateEntry[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label="cron update"
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          changed: r.changed,
          error: r.error,
          extra: r.name ? <Text variant="muted">{r.name}</Text> : undefined,
        }))}
        data={data}
      />
    );
  }

  // Package list / get / check-updates — expanded with search
  if (
    (type === "package-list" ||
      type === "package-get" ||
      type === "package-check-updates") &&
    Array.isArray(data.results)
  ) {
    const pkgLabel =
      type === "package-list"
        ? "package list"
        : type === "package-get"
          ? "package get"
          : "available updates";
    return <PackageListResult data={data} label={pkgLabel} />;
  }

  // Package install / remove / update — mutation results
  if (
    (type === "package-install" ||
      type === "package-remove" ||
      type === "package-update") &&
    Array.isArray(data.results)
  ) {
    type PackageMutationResult = {
      hostname: string;
      status?: string;
      name?: string;
      changed?: boolean;
      error?: string;
    };
    const results = data.results as PackageMutationResult[];
    const label =
      type === "package-install"
        ? "install"
        : type === "package-remove"
          ? "remove"
          : "update";
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label={label}
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          changed: r.changed,
          error: r.error,
          extra: r.name ? <Text variant="muted">{r.name}</Text> : undefined,
        }))}
        data={data}
      />
    );
  }

  // Sysctl list — expanded with search
  if (type === "sysctl-list" && Array.isArray(data.results)) {
    return <SysctlListResult data={data} />;
  }

  // Config operations (sysctl, ntp, timezone, hostname)
  if (
    [
      "sysctl-get",
      "sysctl-set",
      "sysctl-update",
      "sysctl-delete",
      "ntp-get",
      "ntp-set",
      "ntp-update",
      "ntp-delete",
      "timezone-get",
      "timezone-set",
      "hostname-get",
      "hostname-set",
    ].includes(type) &&
    Array.isArray(data.results)
  ) {
    const label = type.replace("-", " ");
    const results = data.results as {
      hostname?: string;
      status?: string;
      changed?: boolean;
      error?: string;
    }[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label={label}
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          changed: r.changed,
          error: r.error,
        }))}
        data={data}
      />
    );
  }

  // Process list — expanded with search
  if (type === "process-list" && Array.isArray(data.results)) {
    return <ProcessListResult data={data} />;
  }

  // Log sources — expanded with search
  if (type === "log-sources" && Array.isArray(data.results)) {
    return <LogSourcesResult data={data} />;
  }

  // System operations (power, process, log)
  if (
    [
      "power-reboot",
      "power-shutdown",
      "process-get",
      "process-signal",
      "log-query",
      "log-query-unit",
    ].includes(type) &&
    Array.isArray(data.results)
  ) {
    const label = type.replace("-", " ");
    const results = data.results as {
      hostname?: string;
      status?: string;
      changed?: boolean;
      error?: string;
    }[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label={label}
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          changed: r.changed,
          error: r.error,
        }))}
        data={data}
      />
    );
  }

  // User list — expanded with search
  if (type === "user-list" && Array.isArray(data.results)) {
    return <UserListResult data={data} />;
  }

  // Group list — expanded with search
  if (type === "group-list" && Array.isArray(data.results)) {
    return <GroupListResult data={data} />;
  }

  // Certificate list — expanded with search
  if (type === "certificate-list" && Array.isArray(data.results)) {
    return <CertificateListResult data={data} />;
  }

  // Security — user, group, certificate operations
  if (
    [
      "user-get",
      "user-create",
      "user-update",
      "user-delete",
      "user-list-keys",
      "user-add-key",
      "user-remove-key",
      "user-change-password",
      "group-get",
      "group-create",
      "group-update",
      "group-delete",
      "certificate-create",
      "certificate-update",
      "certificate-delete",
    ].includes(type) &&
    Array.isArray(data.results)
  ) {
    const label = type.replace("-", " ");
    const results = data.results as {
      hostname?: string;
      status?: string;
      changed?: boolean;
      error?: string;
    }[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label={label}
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          changed: r.changed,
          error: r.error,
        }))}
        data={data}
      />
    );
  }

  // Docker inspect — per-host container detail
  if (type === "docker-inspect" && Array.isArray(data.results)) {
    type DockerDetailEntry = {
      hostname: string;
      status?: string;
      id?: string;
      name?: string;
      image?: string;
      state?: string;
      error?: string;
    };
    const results = data.results as DockerDetailEntry[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label="inspect"
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          error: r.error,
          extra: (
            <span className="flex items-center gap-1.5">
              {r.name && <Text>{r.name}</Text>}
              {r.image && <Text variant="muted">{r.image}</Text>}
              {r.state && (
                <Text variant={r.state === "running" ? "primary" : "muted"}>
                  {r.state}
                </Text>
              )}
            </span>
          ),
        }))}
        data={data}
      />
    );
  }

  // File stale — list of stale deployments
  if (type === "file-stale" && Array.isArray(data.stale)) {
    type StaleEntry = {
      object_name: string;
      hostname: string;
      provider: string;
      path: string;
      deployed_sha: string;
      current_sha: string;
      deployed_at: string;
    };
    const entries = data.stale as StaleEntry[];
    return (
      <ResultShell
        label={`${entries.length} stale deployment${entries.length !== 1 ? "s" : ""}`}
        summary={
          <div
            className={`h-1.5 w-1.5 rounded-full ${entries.length > 0 ? "bg-status-error" : "bg-primary"}`}
          />
        }
        data={data}
      >
        {entries.map((e, i) => (
          <div
            key={i}
            className="flex items-center gap-4 border-b border-border/50 px-4 py-2 last:border-0"
          >
            <StatusIcon status="failed" />
            <KeyValue label="host" value={e.hostname} variant="strong" />
            <KeyValue label="object" value={e.object_name} variant="accent" />
            <KeyValue label="path" value={e.path} />
            <Text variant="mono-muted" className="ml-auto">
              sha:{e.current_sha.slice(0, 12)}
            </Text>
          </div>
        ))}
        {entries.length === 0 && (
          <div className="px-4 py-3">
            <Text variant="muted">No stale deployments found.</Text>
          </div>
        )}
      </ResultShell>
    );
  }

  // Node status — per-host with uptime, load, memory, disk, os_info
  if (type === "node-status" && Array.isArray(data.results)) {
    type NodeStatusRow = {
      hostname: string;
      status?: string;
      uptime?: string;
      load_average?: { "1min": number; "5min": number; "15min": number };
      memory?: { total: number; used: number; free: number };
      disks?: { name: string; total: number; used: number; free: number }[];
      os_info?: { distribution: string; version: string };
      error?: string;
    };
    const results = data.results as NodeStatusRow[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label="node status"
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          error: r.error,
          extra: !r.error ? (
            <div className="flex flex-wrap gap-3 text-xs text-text-muted">
              {r.uptime && <span>up {r.uptime}</span>}
              {r.load_average && (
                <span>
                  load {r.load_average["1min"].toFixed(2)}/
                  {r.load_average["5min"].toFixed(2)}/
                  {r.load_average["15min"].toFixed(2)}
                </span>
              )}
              {r.memory && (
                <span>
                  mem {formatSize(r.memory.used)}/{formatSize(r.memory.total)}
                </span>
              )}
              {r.os_info && (
                <span>
                  {r.os_info.distribution} {r.os_info.version}
                </span>
              )}
            </div>
          ) : undefined,
        }))}
        data={data}
      />
    );
  }

  // Node load — per-host load averages
  if (type === "node-load" && Array.isArray(data.results)) {
    type LoadRow = {
      hostname: string;
      status?: string;
      load_average?: { "1min": number; "5min": number; "15min": number };
      error?: string;
    };
    const results = data.results as LoadRow[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label="load"
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          error: r.error,
          extra:
            r.load_average && !r.error ? (
              <Text variant="muted">
                {r.load_average["1min"].toFixed(2)} /{" "}
                {r.load_average["5min"].toFixed(2)} /{" "}
                {r.load_average["15min"].toFixed(2)}
              </Text>
            ) : undefined,
        }))}
        data={data}
      />
    );
  }

  // Node uptime — per-host uptime string
  if (type === "node-uptime" && Array.isArray(data.results)) {
    type UptimeRow = {
      hostname: string;
      status?: string;
      uptime?: string;
      error?: string;
    };
    const results = data.results as UptimeRow[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label="uptime"
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          error: r.error,
          extra:
            r.uptime && !r.error ? (
              <Text variant="muted">{r.uptime}</Text>
            ) : undefined,
        }))}
        data={data}
      />
    );
  }

  // Node OS — per-host distribution and version
  if (type === "node-os" && Array.isArray(data.results)) {
    type OsRow = {
      hostname: string;
      status?: string;
      os_info?: { distribution: string; version: string };
      error?: string;
    };
    const results = data.results as OsRow[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label="os info"
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          error: r.error,
          extra:
            r.os_info && !r.error ? (
              <Text variant="muted">
                {r.os_info.distribution} {r.os_info.version}
              </Text>
            ) : undefined,
        }))}
        data={data}
      />
    );
  }

  // Disk info — per-host disk list
  if (type === "disk-info" && Array.isArray(data.results)) {
    type DiskRow = {
      hostname: string;
      status?: string;
      disks?: { name: string; total: number; used: number; free: number }[];
      error?: string;
    };
    const results = data.results as DiskRow[];
    return (
      <ResultShell
        jobId={data.job_id as string | undefined}
        label={`${results.length} host${results.length !== 1 ? "s" : ""} · disk`}
        summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
        data={data}
      >
        {results.map((r, i) => (
          <div key={i} className="border-b border-border/50 last:border-0">
            <HostGroupHeader
              hostname={r.hostname}
              status={hostStatus(r)}
              detail={
                r.error
                  ? undefined
                  : `${r.disks?.length ?? 0} disk${(r.disks?.length ?? 0) !== 1 ? "s" : ""}`
              }
            />
            {r.error && (
              <Text variant="error" as="p" className="px-4 pb-2">
                {r.error}
              </Text>
            )}
            {!r.error &&
              r.disks?.map((d, j) => (
                <div
                  key={j}
                  className="flex items-center gap-4 border-b border-border/30 px-4 py-2 pl-9 last:border-0"
                >
                  <KeyValue label="dev" value={d.name} variant="strong" />
                  <KeyValue label="total" value={formatSize(d.total)} />
                  <KeyValue label="used" value={formatSize(d.used)} />
                  <KeyValue label="free" value={formatSize(d.free)} />
                </div>
              ))}
          </div>
        ))}
      </ResultShell>
    );
  }

  // Memory info — per-host memory stats
  if (type === "memory-info" && Array.isArray(data.results)) {
    type MemRow = {
      hostname: string;
      status?: string;
      memory?: { total: number; used: number; free: number };
      error?: string;
    };
    const results = data.results as MemRow[];
    return (
      <CollectionResult
        jobId={data.job_id as string | undefined}
        label="memory"
        results={results.map((r) => ({
          hostname: r.hostname,
          status: hostStatus(r),
          error: r.error,
          extra:
            r.memory && !r.error ? (
              <div className="flex gap-3 text-xs text-text-muted">
                <span>total {formatSize(r.memory.total)}</span>
                <span>used {formatSize(r.memory.used)}</span>
                <span>free {formatSize(r.memory.free)}</span>
              </div>
            ) : undefined,
        }))}
        data={data}
      />
    );
  }

  // Audit list / export — tabular list of entries
  if (
    (type === "audit-list" || type === "audit-export") &&
    "items" in data &&
    Array.isArray(data.items)
  ) {
    type AuditRow = {
      id: string;
      timestamp: string;
      user: string;
      method: string;
      path: string;
      response_code: number;
      duration_ms: number;
      source_ip: string;
    };
    const items = data.items as AuditRow[];
    const total = data.total_items as number | undefined;
    return (
      <ResultShell
        label={`${total ?? items.length} entr${(total ?? items.length) !== 1 ? "ies" : "y"}`}
        summary={<div className="h-1.5 w-1.5 rounded-full bg-primary" />}
        data={data}
      >
        {items.map((entry, i) => (
          <div
            key={i}
            className="flex items-center gap-3 border-b border-border/50 px-4 py-2 last:border-0"
          >
            <Text variant="mono-muted" className="shrink-0">
              {new Date(entry.timestamp).toLocaleString()}
            </Text>
            <Text variant="accent" className="shrink-0 font-mono">
              {entry.method}
            </Text>
            <Text className="min-w-0 truncate font-mono">{entry.path}</Text>
            <Text
              variant={entry.response_code < 400 ? "primary" : "error"}
              className="ml-auto shrink-0 font-mono"
            >
              {entry.response_code}
            </Text>
            <Text variant="muted" className="shrink-0">
              {entry.duration_ms}ms
            </Text>
          </div>
        ))}
        {items.length === 0 && (
          <div className="px-4 py-3">
            <Text variant="muted">No audit entries found.</Text>
          </div>
        )}
      </ResultShell>
    );
  }

  // Audit get — single entry
  if (type === "audit-get" && "entry" in data) {
    type AuditEntry = {
      id: string;
      timestamp: string;
      user: string;
      roles: string[];
      method: string;
      path: string;
      operation_id?: string;
      source_ip: string;
      response_code: number;
      duration_ms: number;
      trace_id?: string;
    };
    const entry = data.entry as AuditEntry;
    return (
      <ResultShell label="audit entry" summary={null} data={data}>
        <div className="space-y-1 px-4 py-3">
          <KeyValue label="id" value={entry.id} />
          <KeyValue
            label="timestamp"
            value={new Date(entry.timestamp).toLocaleString()}
          />
          <KeyValue label="user" value={entry.user} />
          <KeyValue label="roles" value={entry.roles.join(", ")} />
          <KeyValue label="method" value={entry.method} />
          <KeyValue label="path" value={entry.path} />
          {entry.operation_id && (
            <KeyValue label="operation" value={entry.operation_id} />
          )}
          <KeyValue label="source_ip" value={entry.source_ip} />
          <KeyValue label="status" value={String(entry.response_code)} />
          <KeyValue label="duration" value={`${entry.duration_ms}ms`} />
          {entry.trace_id && (
            <KeyValue label="trace_id" value={entry.trace_id} />
          )}
        </div>
      </ResultShell>
    );
  }

  // Fallback — unknown type, show in shell with JSON toggle
  return <ResultShell label="result" summary={null} data={data} />;
}

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
