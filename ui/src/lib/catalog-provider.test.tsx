import { afterEach, describe, expect, mock, test } from "bun:test";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { CatalogProvider } from "@/lib/catalog-provider";
import { SessionProvider } from "@/lib/session";
import { useCatalog } from "@/hooks/use-catalog";
import { useStatus } from "@/hooks/use-status";
import { testCatalog } from "@/test/catalog";

const realFetch = globalThis.fetch;

afterEach(() => {
  globalThis.fetch = realFetch;
});

const answer = (status: number, body: unknown) => {
  const fetch = mock(
    async () => new Response(JSON.stringify(body), { status }),
  );
  globalThis.fetch = fetch as unknown as typeof globalThis.fetch;
  return fetch;
};

function Probe() {
  const { roles } = useCatalog();
  const status = useStatus();
  return (
    <p>
      {roles.map((r) => r.name).join(",")} {status.variant("failed")}{" "}
      {status.dot("skipped")} {status.text("ok")} {status.variant("other")}
    </p>
  );
}

const app = () =>
  render(
    <SessionProvider>
      <CatalogProvider>
        <Probe />
      </CatalogProvider>
    </SessionProvider>,
  );

describe("CatalogProvider", () => {
  test("loads the catalog once, then shows the pages with it", async () => {
    const fetch = answer(200, testCatalog);
    app();
    await waitFor(() =>
      expect(
        screen.getByText("admin,read,observer error warn primary muted"),
      ).toBeTruthy(),
    );
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(String(fetch.mock.calls[0])).toContain("/api/catalog");
  });

  test("says why when the load fails, with a retry and a sign out", async () => {
    answer(403, { error: "missing permission catalog:read" });
    app();
    await waitFor(() =>
      expect(
        screen.getByText(
          "Could not load the catalog: missing permission catalog:read",
        ),
      ).toBeTruthy(),
    );
    expect(screen.getByText("Sign out")).toBeTruthy();

    answer(200, testCatalog);
    fireEvent.click(screen.getByText("Retry"));
    await waitFor(() =>
      expect(screen.getByText(/admin,read,observer/)).toBeTruthy(),
    );
  });
});

describe("useCatalog", () => {
  test("refuses a page outside CatalogProvider", () => {
    const quiet = mock(() => {});
    const error = console.error;
    console.error = quiet;
    expect(() => render(<Probe />)).toThrow(/CatalogProvider/);
    console.error = error;
  });
});
