import { describe, expect, test } from "bun:test";
import { certSummary } from "@/lib/certs";
import type { PKICert } from "@/sdk/gen/schemas";

const cert = (
  name: string,
  hostname: string,
  days?: number,
  state = "ok",
): PKICert => ({
  name,
  hostname,
  state,
  component: "agent",
  source: "self-signed",
  file: `/${name}.crt`,
  days_left: days,
});

describe("certSummary", () => {
  test("the certificate that expires first, of the ones that osapi read", () => {
    const got = certSummary([
      cert("api", "hub", 78),
      cert("agent-metrics", "web-01", 1824),
      cert("nats", "hub", 12),
      cert("nats-metrics", "hub", undefined, "not made"),
    ]);
    expect(got.soonest?.name).toBe("nats");
    expect(got.ok).toBe(3);
    expect(got.total).toBe(4);
  });

  test("no certificate that osapi read has none first", () => {
    const got = certSummary([
      cert("nats-metrics", "hub", undefined, "not made"),
    ]);
    expect(got.soonest).toBeUndefined();
    expect(got.ok).toBe(0);
  });
});
