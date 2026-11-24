import { useState } from "react";
import { KeyRound, Trash2 } from "lucide-react";
import { unwrap, usePolling } from "@/hooks/use-polling";
import { useAuth } from "@/hooks/use-auth";
import { formatTime } from "@/lib/format";
import { ContentArea } from "@/components/layout/content-area";
import { PermissionDenied } from "@/components/layout/permission-denied";
import { PageHeader } from "@/components/ui/page-header";
import { Picker } from "@/components/ui/popover";
import { DataTable } from "@/components/ui/data-table";
import { EmptyState } from "@/components/ui/empty-state";
import { ErrorBanner } from "@/components/ui/error-banner";
import { SectionLabel } from "@/components/ui/section-label";
import { Spinner } from "@/components/ui/spinner";
import { IconButton } from "@/components/ui/icon-button";
import { ConfirmModal } from "@/components/ui/confirm-modal";
import { Text } from "@/components/ui/text";
import { SecretForm } from "@/components/domain/secret-form";
import type {
  SecretInfo,
  SecretListResponse,
  VaultInfo,
  VaultListResponse,
} from "@/sdk/gen/schemas";
import {
  deleteSecret,
  getSecretVaults,
  getSecrets,
  putSecret,
} from "@/sdk/gen/secret-management-api-secret-operations/secret-management-api-secret-operations";

/**
 * Secrets lists the keys of a vault, with the columns of osapi client
 * secret list, writes a value and deletes a key. No API answer has a
 * value, so the page has none to show.
 */
export function Secrets() {
  const { canCall } = useAuth();
  const [picked, setPicked] = useState("");
  const [actionError, setActionError] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);

  const vaults = usePolling(
    async () =>
      unwrap<VaultListResponse>(
        await getSecretVaults(),
        "Failed to fetch the vaults",
      ).items ?? [],
    [],
    { enabled: canCall("GetSecretVaults") },
  );
  const vaultList: VaultInfo[] = vaults.data ?? [];
  const vault = picked || vaultList[0]?.name || "";

  const keys = usePolling(
    async () =>
      vault === ""
        ? []
        : (unwrap<SecretListResponse>(
            await getSecrets(vault),
            "Failed to fetch the keys of the vault",
          ).items ?? []),
    [vault],
    { enabled: canCall("GetSecrets") },
  );
  const rows: SecretInfo[] = keys.data ?? [];

  if (!canCall("GetSecretVaults")) {
    return <PermissionDenied title="Secrets" what="the vaults" />;
  }

  const write = async (key: string, value: string) => {
    setActionError(null);
    const res = await putSecret(vault, key, { value });
    if (res.status !== 200) {
      setActionError(
        (res.data as { error?: string } | undefined)?.error ??
          "Failed to write the key",
      );
      return false;
    }
    await keys.refresh();
    return true;
  };

  const remove = async (key: string) => {
    setDeleting(true);
    setActionError(null);
    const res = await deleteSecret(vault, key);
    if (res.status !== 204) {
      setActionError(
        (res.data as { error?: string } | undefined)?.error ??
          "Failed to delete the key",
      );
    }
    setDeleting(false);
    setConfirmDelete(null);
    await keys.refresh();
  };

  const selected = vaultList.find((v) => v.name === vault);

  return (
    <ContentArea>
      <PageHeader
        title="Secrets"
        subtitle="The keys of the vaults of the controller. No page shows a value."
      />

      {vaults.loading && <Spinner />}
      {vaults.error && <ErrorBanner message={vaults.error} />}
      {actionError && <ErrorBanner message={actionError} />}

      {!vaults.loading && !vaults.error && vaultList.length === 0 && (
        <EmptyState
          icon={<KeyRound className="h-8 w-8 text-text-muted" />}
          message="The controller has no vault. Add one to controller.secrets.vaults."
        />
      )}

      {vaultList.length > 0 && (
        <>
          <div className="mb-6 max-w-sm">
            <Picker
              id="secret-vault"
              label="Vault"
              value={vault}
              onChange={setPicked}
              placeholder="Pick a vault"
              options={vaultList.map((v) => ({
                value: v.name,
                meta: v.answers ? v.type : `${v.type}, does not answer`,
              }))}
            >
              {selected && (
                <Text variant="muted" as="p" className="mt-1 text-xs">
                  Targets:{" "}
                  {selected.targets.length > 0
                    ? selected.targets.join(", ")
                    : "each host"}
                </Text>
              )}
            </Picker>
          </div>

          {canCall("PutSecret") && (
            <section className="mb-6">
              <SectionLabel>Write a value</SectionLabel>
              <SecretForm key={vault} onWrite={write} />
            </section>
          )}

          <SectionLabel>Keys</SectionLabel>
          {keys.error && <ErrorBanner message={keys.error} />}
          <DataTable
            columns={[
              {
                header: "Key",
                cell: (r: SecretInfo) => <Text variant="mono">{r.key}</Text>,
              },
              { header: "Version", cell: (r: SecretInfo) => r.version },
              {
                header: "Changed",
                cell: (r: SecretInfo) => formatTime(r.changed),
              },
              { header: "By", cell: (r: SecretInfo) => r.by || "-" },
              ...(canCall("DeleteSecret")
                ? [
                    {
                      header: "",
                      align: "right" as const,
                      cell: (r: SecretInfo) => (
                        <IconButton
                          icon={Trash2}
                          variant="danger"
                          title="Delete key"
                          onClick={() => setConfirmDelete(r.key)}
                        />
                      ),
                    },
                  ]
                : []),
            ]}
            rows={rows}
            getRowKey={(r) => r.key}
          />
        </>
      )}

      <ConfirmModal
        open={confirmDelete !== null}
        title="Delete key"
        action="Delete"
        busy={deleting}
        onConfirm={() => confirmDelete && remove(confirmDelete)}
        onClose={() => setConfirmDelete(null)}
      >
        Delete key <Text variant="mono">{confirmDelete}</Text> from vault{" "}
        <Text variant="mono">{vault}</Text>? A request that uses its reference
        then fails.
      </ConfirmModal>
    </ContentArea>
  );
}
