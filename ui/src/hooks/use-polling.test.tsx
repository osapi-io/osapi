import { describe, expect, mock, test } from "bun:test";
import { renderHook, waitFor } from "@testing-library/react";
import { unwrap, usePolling } from "@/hooks/use-polling";

describe("usePolling", () => {
  test("it calls, and gives the answer", async () => {
    const fetch = mock(async () => 42);
    const { result } = renderHook(() => usePolling(fetch, []));
    expect(result.current.loading).toBe(true);
    await waitFor(() => expect(result.current.data).toBe(42));
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  test("a refused role sends no request", async () => {
    const fetch = mock(async () => 1);
    const { result } = renderHook(() =>
      usePolling(fetch, [], { enabled: false }),
    );
    await new Promise((r) => setTimeout(r, 20));
    expect(fetch).not.toHaveBeenCalled();
    expect(result.current.loading).toBe(false);
  });

  test("a failed call gives its message", async () => {
    const { result } = renderHook(() =>
      usePolling(async () => {
        throw new Error("down");
      }, []),
    );
    await waitFor(() => expect(result.current.error).toBe("down"));
  });

  test("new inputs call again", async () => {
    const fetch = mock(async (n: number) => n * 2);
    const { result, rerender } = renderHook(
      ({ n }) => usePolling(() => fetch(n), [n], { intervalMs: 0 }),
      { initialProps: { n: 1 } },
    );
    await waitFor(() => expect(result.current.data).toBe(2));
    rerender({ n: 2 });
    await waitFor(() => expect(result.current.data).toBe(4));
  });
});

describe("unwrap", () => {
  test("the body of the expected status", () => {
    expect(unwrap<number>({ status: 200, data: 7 }, "x")).toBe(7);
  });

  test("the API's error, or the fallback", () => {
    expect(() =>
      unwrap({ status: 403, data: { error: "forbidden" } }, "x"),
    ).toThrow("forbidden");
    expect(() => unwrap({ status: 500, data: undefined }, "x")).toThrow("x");
  });
});
