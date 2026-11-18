// formatUptime shortens the API's uptime, "2 days, 23 hours, 50 minutes", to
// the two largest units, "2d 23h", as the CLI's FormatAge does. A string in a
// different form comes back unchanged.
export function formatUptime(uptime: string): string {
  const unit = (name: string) => {
    const m = uptime.match(new RegExp(`(\\d+)\\s+${name}s?\\b`));
    return m ? Number(m[1]) : undefined;
  };
  const days = unit("day");
  const hours = unit("hour");
  const minutes = unit("minute");
  if (days === undefined && hours === undefined && minutes === undefined) {
    return uptime;
  }

  const d = days ?? 0;
  const h = hours ?? 0;
  const m = minutes ?? 0;
  if (d > 0) return `${d}d ${h}h`;
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}
