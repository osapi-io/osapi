import { traceIdOf } from "@/lib/trace";
import { apiFetch } from "@/sdk/fetch";
import { methodRequest, type CatalogMethod } from "@/lib/catalog";

/**
 * Runs a method against the target, as the generated client would. A
 * refusal (a status of 400 or more) throws the API's error, so the step
 * shows it.
 */
export async function runMethod(
  method: CatalogMethod,
  target: string,
  data: Record<string, unknown>,
): Promise<{ data: unknown }> {
  const { url, init } = methodRequest(method, target, data);
  const res = await apiFetch<{ data: unknown; status: number }>(url, init);
  if (res.status >= 400) {
    const body = res.data as { error?: string; reason?: string } | undefined;
    const reason = body?.reason ? ` (${body.reason})` : "";
    // The error keeps the trace, so the step links it also when it failed.
    throw Object.assign(
      new Error(`${body?.error ?? `HTTP ${res.status}`}${reason}`),
      { traceId: traceIdOf(res) },
    );
  }
  return res;
}
