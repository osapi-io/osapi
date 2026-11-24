import { unwrap, usePolling } from "@/hooks/use-polling";
import { enumLabel } from "@/lib/format";
import { GetAdmissionsState } from "@/sdk/gen/schemas";
import { ContentArea } from "@/components/layout/content-area";
import { AdmissionCard } from "@/components/domain/admission-card";
import { PageHeader } from "@/components/ui/page-header";
import { EmptyState } from "@/components/ui/empty-state";
import { SectionLabel } from "@/components/ui/section-label";
import { useAuth } from "@/hooks/use-auth";
import { ErrorBanner } from "@/components/ui/error-banner";
import { Spinner } from "@/components/ui/spinner";
import { PermissionDenied } from "@/components/layout/permission-denied";
import { ShieldAlert } from "lucide-react";
import type { Admission as Record } from "@/sdk/gen/schemas";
import { getAdmissions } from "@/sdk/gen/admission-api-admission/admission-api-admission";

// The sections, in the order a reader acts on them: what waits for a
// decision first, then what is in, then what was kept out.
const sections = Object.values(GetAdmissionsState).map((state) => ({
  state,
  title: enumLabel(state),
}));

export function Admission() {
  const { canCall } = useAuth();
  // A failed load shows as an error, as on the other pages, not as an
  // empty page that reads as no machines.
  const admissions = usePolling(
    async () =>
      unwrap<{ items?: Record[] }>(
        await getAdmissions(),
        "Failed to fetch admission records",
      ).items ?? [],
    [],
    { enabled: canCall("GetAdmissions") },
  );
  const records = admissions.data ?? [];
  const { error, loading } = admissions;
  const fetchRecords = admissions.refresh;

  if (!canCall("GetAdmissions")) {
    return <PermissionDenied title="Admission" what="admission records" />;
  }

  return (
    <ContentArea>
      <PageHeader
        title="Admission"
        subtitle="Admit, reject and revoke the machines that ask to join"
      />

      {loading && <Spinner />}

      {error && <ErrorBanner message={error} />}

      {!loading && !error && records.length === 0 && (
        <EmptyState
          icon={<ShieldAlert className="h-8 w-8 text-text-muted" />}
          message="No machine has asked to join. Start an agent with PKI on, and it shows here."
        />
      )}

      {sections.map(({ state, title }) => {
        const items = records.filter((r) => r.state === state);
        if (items.length === 0) return null;
        return (
          <section key={state} className="mb-8">
            <SectionLabel>
              {title} ({items.length})
            </SectionLabel>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {items.map((r) => (
                <AdmissionCard
                  key={r.machine_id}
                  record={r}
                  canWrite={canCall("AdmitMachine")}
                  onRefresh={fetchRecords}
                />
              ))}
            </div>
          </section>
        );
      })}
    </ContentArea>
  );
}
