import { describe, expect, test } from "bun:test";
import { formatCell, matches } from "@/lib/result-format";
import type { CatalogColumn } from "@/lib/catalog";

const col = (render?: string): CatalogColumn =>
  ({ name: "v", label: "V", doc: "", kind: "value", render }) as CatalogColumn;

describe("formatCell", () => {
  test("a value as the CLI prints it, by its render tag", () => {
    expect(formatCell(col("short=3"), "abcdef")).toBe("abc...");
    expect(formatCell(col("short=10"), "abc")).toBe("abc");
    expect(formatCell(col(), true)).toBe("yes");
    expect(formatCell(col(), ["a", "b"])).toBe("a, b");
    expect(formatCell(col("bytes"), 2048)).toContain("KB");
  });
});

describe("matches", () => {
  test("a filter term in any field, in any case", () => {
    expect(matches("WEB", "web-01", 3)).toBe(true);
    expect(matches("db", "web-01")).toBe(false);
    expect(matches("", "x")).toBe(true);
  });
});
