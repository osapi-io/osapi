import { describe, expect, test } from "bun:test";
import { fireEvent, render, screen } from "@testing-library/react";
import { TLSSection, TLSSummary } from "@/components/domain/tls-section";
import { formatTime } from "@/lib/format";
import type { PKITLSSection } from "@/sdk/gen/schemas";

const nats: PKITLSSection = {
  component: "nats",
  hostname: "hub",
  summary: "FIPS off · cert 12d",
  fips: "off",
  expiry: "soon",
  reasons: [
    "FIPS 140-3 mode is off: set GODEBUG=fips140=on",
    "FIPS 140-3 only mode refuses JetStream store: nats-server picks its own GCM nonce (osapi#1549)",
  ],
  certificates: [
    {
      name: "nats",
      state: "ok",
      subject: "CN=hub",
      names: ["hub.lab", "10.0.0.1"],
      issuer: "CN=osapi CA",
      source: "self-signed",
      not_before: "2026-01-01T00:00:00Z",
      expires: "2026-12-09T00:00:00Z",
      days_left: 12,
      expiry: "soon",
      key: "ECDSA P-256",
      fips_approved: true,
    },
  ],
  listeners: [
    {
      name: "nats",
      tls: true,
      min_version: "TLS 1.2",
      suites: ["TLS_AES_128_GCM_SHA256"],
      groups: ["CurveP256"],
      approved: true,
    },
  ],
  go: "go1.27.3",
  module: "v1.0.0",
  algorithms: [
    {
      name: "JetStream store",
      algorithm: "AES-256-GCM (nats-server)",
      approved: true,
      only_mode: false,
      reason: "osapi#1549",
    },
  ],
};

describe("TLSSummary", () => {
  test("a certificate that expires soon is pending, and the summary opens the section", () => {
    render(<TLSSummary section={nats} />);
    const link = screen.getByText("FIPS off · cert 12d");
    expect(link.className).toContain("text-status-pending");
    expect(screen.queryByText("Certificates")).toBeNull();
    fireEvent.click(link);
    expect(screen.getByText("Certificates")).toBeDefined();
  });

  test("an expired certificate is an error", () => {
    render(
      <TLSSummary
        section={{
          ...nats,
          expiry: "expired",
          summary: "FIPS off · cert expired",
        }}
      />,
    );
    expect(screen.getByText("FIPS off · cert expired").className).toContain(
      "text-status-error",
    );
  });
});

describe("TLSSection", () => {
  test("it shows the reasons, the certificate, the listener and the algorithms", () => {
    render(<TLSSection section={nats} />);
    expect(screen.getByText("TLS of the nats on hub")).toBeDefined();
    expect(screen.getByText(/refuses JetStream store/)).toBeDefined();
    expect(screen.getByText("hub.lab, 10.0.0.1")).toBeDefined();
    expect(screen.getByText("CN=osapi CA")).toBeDefined();
    expect(
      screen.getByText(
        `${formatTime("2026-01-01T00:00:00Z")} to ${formatTime("2026-12-09T00:00:00Z")}`,
      ),
    ).toBeDefined();
    expect(screen.getByText("12d").className).toContain("text-status-pending");
    expect(screen.getByText("TLS_AES_128_GCM_SHA256")).toBeDefined();
    expect(
      screen.getByText("Go go1.27.3, Go Cryptographic Module v1.0.0"),
    ).toBeDefined();
  });

  test("with nothing to say, each part says so", () => {
    render(
      <TLSSection
        section={{
          component: "agent",
          hostname: "web-01",
          summary: "FIPS unknown",
          fips: "",
          expiry: "unknown",
        }}
      />,
    );
    expect(screen.getByText("Mode: not reported")).toBeDefined();
    expect(
      screen.getByText("Each listener and each algorithm is approved."),
    ).toBeDefined();
    expect(
      screen.getByText("The component serves no certificate that it reports."),
    ).toBeDefined();
  });
});
