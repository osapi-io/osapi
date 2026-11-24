import { describe, expect, test } from "bun:test";
import { render, screen } from "@testing-library/react";
import { AdmissionCard } from "@/components/domain/admission-card";
import type { Admission } from "@/sdk/gen/schemas";

const record: Admission = {
  machine_id: "d530664edc1c431eb5ef1e9dfd6617a5",
  hostname: "web-03",
  state: "pending",
  online: true,
  fingerprint:
    "SHA256:1f53dea6b0c3e2f9a8d4c71e5b6a0f2d9c8e7b6a5f4e3d2c1b0a9f8e7d6c5b4a",
  encryption_key: "XA7NTLWJ2ECILFDLHORNVYROXOK65ONQ2SJYHRNAWRJZ3UXMVXIGUVJM",
};

describe("AdmissionCard", () => {
  test("shows the fingerprint and the encryption key that it covers", () => {
    render(<AdmissionCard record={record} canWrite={false} />);
    expect(screen.getByText("Fingerprint")).toBeTruthy();
    expect(screen.getByText("Encryption key")).toBeTruthy();
    expect(screen.getByText("XA7NTLWJ2ECILFDLHORN...")).toBeTruthy();
  });

  test("a record with no encryption key has no encryption key row", () => {
    render(
      <AdmissionCard
        record={{ ...record, encryption_key: undefined }}
        canWrite={false}
      />,
    );
    expect(screen.queryByText("Encryption key")).toBeNull();
  });
});
