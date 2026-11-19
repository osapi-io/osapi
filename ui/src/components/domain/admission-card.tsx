import { useState } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { CopyField } from "@/components/ui/copy-field";
import { Input } from "@/components/ui/input";
import { Text } from "@/components/ui/text";
import { HealthDot } from "@/components/ui/health-dot";
import {
  Check,
  X,
  Loader2,
  ShieldAlert,
  ShieldCheck,
  ShieldOff,
  Trash2,
} from "lucide-react";
import type { Admission } from "@/sdk/gen/schemas";
import {
  admitMachine,
  rejectMachine,
  revokeMachine,
  deleteMachine,
} from "@/sdk/gen/admission-api-admission/admission-api-admission";
import { statusVariant } from "@/lib/status";
import { formatAgeSince, shortFingerprint } from "@/lib/format";

interface AdmissionCardProps {
  record: Admission;
  /** The viewer has admission:write. */
  canWrite: boolean;
  onRefresh?: () => void;
}

// Which form is open: admit asks for the fingerprint that the machine
// prints, reject for a reason, revoke for a second click.
type Form = "admit" | "reject" | "revoke" | "delete" | null;

const icons = {
  pending: ShieldAlert,
  admitted: ShieldCheck,
} as const;

export function AdmissionCard({
  record,
  canWrite,
  onRefresh,
}: AdmissionCardProps) {
  const [form, setForm] = useState<Form>(null);
  const [value, setValue] = useState("");
  const [acting, setActing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const Icon = icons[record.state as keyof typeof icons] ?? ShieldOff;

  const open = (f: Form) => {
    setForm(f);
    setValue("");
    setError(null);
  };

  const submit = async () => {
    setActing(true);
    setError(null);
    try {
      const resp =
        form === "admit"
          ? await admitMachine(record.machine_id, { fingerprint: value.trim() })
          : form === "reject"
            ? await rejectMachine(record.machine_id, { reason: value.trim() })
            : form === "revoke"
              ? await revokeMachine(record.machine_id)
              : await deleteMachine(record.machine_id);
      if (resp.status === 204) {
        setForm(null);
        onRefresh?.();
      } else {
        setError((resp.data as { error?: string })?.error ?? `${form} failed`);
      }
    } catch {
      setError(`${form} failed`);
    }
    setActing(false);
  };

  return (
    <Card variant="active" className="flex flex-col">
      <CardHeader>
        <Icon className="h-4 w-4 text-accent-light" />
        <CardTitle className="truncate" title={record.hostname}>
          {record.hostname}
        </CardTitle>
        <div className="ml-auto flex items-center gap-2">
          <span title={record.online ? "Online" : "Offline"}>
            <HealthDot status={record.online} />
          </span>
          <Badge variant={statusVariant(record.state)}>{record.state}</Badge>
        </div>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-1.5">
        <div className="space-y-0.5">
          <CopyField label="Machine ID" value={record.machine_id} />
          <CopyField
            label="Fingerprint"
            value={record.fingerprint}
            display={shortFingerprint(record.fingerprint)}
          />
        </div>
        <Text variant="muted" as="p">
          Last seen {formatAgeSince(record.last_seen) || "never"}
          {record.requested_at &&
            ` · requested ${formatAgeSince(record.requested_at)}`}
        </Text>
        {record.reason && (
          <Text variant="muted" as="p">
            Reason: {record.reason}
          </Text>
        )}
        {record.previous && (
          <Text variant="warn" as="p">
            Asks again after it was {record.previous}
            {record.previous_fingerprint &&
              record.previous_fingerprint !== record.fingerprint &&
              ", with a different key"}
            .
          </Text>
        )}

        {form && form !== "revoke" && (
          <Input
            autoFocus
            placeholder={
              form === "admit"
                ? "Fingerprint from osapi agent admission"
                : "Reason (optional)"
            }
            value={value}
            onChange={(e) => setValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") submit();
              if (e.key === "Escape") setForm(null);
            }}
          />
        )}
        {error && (
          <Text variant="error" as="p">
            {error}
          </Text>
        )}

        {canWrite && (
          <div className="mt-auto flex gap-2 border-t border-border/30 pt-2">
            {form ? (
              <>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={submit}
                  disabled={acting || (form === "admit" && !value.trim())}
                  className={
                    form === "admit" ? "text-primary" : "text-status-error"
                  }
                >
                  {acting ? (
                    <Loader2 className="h-3 w-3 animate-spin" />
                  ) : (
                    <Check className="h-3 w-3" />
                  )}
                  {form === "admit"
                    ? "Admit"
                    : form === "reject"
                      ? "Reject"
                      : form === "revoke"
                        ? "Confirm revoke"
                        : "Confirm delete"}
                </Button>
                <Button variant="ghost" size="sm" onClick={() => setForm(null)}>
                  Cancel
                </Button>
              </>
            ) : record.state === "pending" ? (
              <>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => open("admit")}
                  className="text-primary"
                >
                  <Check className="h-3 w-3" />
                  Admit
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => open("reject")}
                >
                  <X className="h-3 w-3" />
                  Reject
                </Button>
              </>
            ) : (
              <>
                {record.state === "admitted" && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => open("revoke")}
                  >
                    <ShieldOff className="h-3 w-3" />
                    Revoke
                  </Button>
                )}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => open("delete")}
                  title="Remove the machine and each record of it, such as before a rebuild"
                >
                  <Trash2 className="h-3 w-3" />
                  Delete
                </Button>
              </>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
