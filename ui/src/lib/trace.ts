// traceIdOf is the trace of an answer of the API, from its W3C traceresponse
// header (version-traceid-spanid-flags), or undefined when it has none. The
// controller sets the header on each answer, as the CLI reads it for the
// trace line of each command.
export function traceIdOf(res: unknown): string | undefined {
  const headers = (res as { headers?: Headers } | undefined)?.headers;
  const value = headers?.get?.("traceresponse");
  if (!value) return undefined;
  const parts = value.split("-");
  if (parts.length !== 4 || !/^[0-9a-f]{32}$/.test(parts[1])) return undefined;
  if (/^0+$/.test(parts[1])) return undefined;
  return parts[1];
}
