import { useState } from "react";
import { Modal } from "@/components/ui/modal";
import { DataTable } from "@/components/ui/data-table";
import { SectionLabel } from "@/components/ui/section-label";
import { formatTime } from "@/lib/format";
import type {
  CryptoAlgorithm,
  CryptoListener,
  PKITLSSection,
  PKITLSCert,
} from "@/sdk/gen/schemas";

// tone is the class of an expiry or a verdict: an error once a certificate
// expired, pending when it is soon or something is not approved, muted when
// all is well.
function tone(section: PKITLSSection): string {
  if (section.expiry === "expired") return "text-status-error";
  if (section.expiry === "soon" || section.fips.includes("not approved")) {
    return "text-status-pending";
  }
  return "text-text-muted";
}

function expiryTone(cert: PKITLSCert): string {
  if (cert.expiry === "expired") return "text-status-error";
  if (cert.expiry === "soon") return "text-status-pending";
  return "";
}

function yesNo(value?: boolean): string {
  if (value === undefined) return "-";
  return value ? "yes" : "no";
}

// TLSSummary is a component's TLS in one line, such as FIPS off · cert 62d.
// It opens the component's TLS section.
export function TLSSummary({ section }: { section: PKITLSSection }) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <button
        type="button"
        className={`mt-0.5 truncate text-left text-[10px] underline decoration-dotted underline-offset-2 hover:text-text ${tone(section)}`}
        onClick={() => setOpen(true)}
        title="Show the TLS of this component"
      >
        {section.summary}
      </button>
      <Modal
        open={open}
        onClose={() => setOpen(false)}
        className="max-h-[85vh] max-w-4xl overflow-y-auto"
      >
        <TLSSection section={section} />
      </Modal>
    </>
  );
}

// TLSSection is the TLS of one component, as osapi client pki tls get
// prints it: FIPS with its reasons, the certificates, the listeners, and the
// versions and algorithms.
export function TLSSection({ section }: { section: PKITLSSection }) {
  return (
    <div className="space-y-4 text-xs">
      <div>
        <h2 className="text-sm font-medium text-text">
          TLS of the {section.component} on {section.hostname}
        </h2>
        <p className={`mt-1 ${tone(section)}`}>{section.summary}</p>
      </div>

      <div>
        <SectionLabel>FIPS 140-3</SectionLabel>
        <p className="text-text">Mode: {section.fips || "not reported"}</p>
        {section.reasons && section.reasons.length > 0 ? (
          <ul className="mt-1 list-disc space-y-0.5 pl-4 text-text-muted">
            {section.reasons.map((r) => (
              <li key={r}>{r}</li>
            ))}
          </ul>
        ) : (
          <p className="mt-1 text-text-muted">
            Each listener and each algorithm is approved.
          </p>
        )}
      </div>

      <div>
        <SectionLabel>Certificates</SectionLabel>
        <DataTable<PKITLSCert>
          rows={section.certificates ?? []}
          getRowKey={(c) => c.name}
          empty="The component serves no certificate that it reports."
          columns={[
            { header: "Name", cell: (c) => c.name },
            {
              header: "Subject and SAN",
              cell: (c) => (
                <span>
                  {c.subject || "-"}
                  {c.names && c.names.length > 0 && (
                    <span className="block text-text-muted">
                      {c.names.join(", ")}
                    </span>
                  )}
                </span>
              ),
            },
            { header: "Issuer", cell: (c) => c.issuer || "-" },
            {
              header: "Valid",
              cell: (c) =>
                `${formatTime(c.not_before) || "-"} to ${formatTime(c.expires) || "-"}`,
            },
            {
              header: "Left",
              cell: (c) => (
                <span className={expiryTone(c)}>
                  {c.expiry === "expired"
                    ? "expired"
                    : c.days_left !== undefined
                      ? `${c.days_left}d`
                      : c.state}
                </span>
              ),
            },
            { header: "Key", cell: (c) => c.key || "-" },
            { header: "Approved", cell: (c) => yesNo(c.fips_approved) },
          ]}
        />
      </div>

      <div>
        <SectionLabel>Listeners</SectionLabel>
        <DataTable<CryptoListener>
          rows={section.listeners ?? []}
          getRowKey={(l) => l.name}
          empty="The component has no TLS listener."
          columns={[
            { header: "Name", cell: (l) => l.name },
            {
              header: "TLS",
              cell: (l) => (l.tls ? (l.min_version ?? "yes") : "no TLS"),
            },
            {
              header: "Suites",
              cell: (l) => (
                <span className="font-mono text-[10px]">
                  {(l.suites ?? []).join(", ") || "-"}
                </span>
              ),
            },
            {
              header: "Groups",
              cell: (l) => (l.groups ?? []).join(", ") || "-",
            },
            {
              header: "Approved",
              cell: (l) => (
                <span className={l.approved ? "" : "text-status-pending"}>
                  {yesNo(l.approved)}
                </span>
              ),
            },
          ]}
        />
      </div>

      <div>
        <SectionLabel>Versions and algorithms</SectionLabel>
        <p className="mb-2 text-text-muted">
          Go {section.go || "-"}, Go Cryptographic Module{" "}
          {section.module || "-"}
        </p>
        <DataTable<CryptoAlgorithm>
          rows={section.algorithms ?? []}
          getRowKey={(a) => a.name}
          empty="The component reports no algorithm."
          columns={[
            { header: "Use", cell: (a) => a.name },
            { header: "Algorithm", cell: (a) => a.algorithm },
            {
              header: "Approved",
              cell: (a) => (
                <span className={a.approved ? "" : "text-status-pending"}>
                  {yesNo(a.approved)}
                </span>
              ),
            },
            {
              header: "Only mode",
              cell: (a) => <span title={a.reason}>{yesNo(a.only_mode)}</span>,
            },
          ]}
        />
      </div>
    </div>
  );
}
