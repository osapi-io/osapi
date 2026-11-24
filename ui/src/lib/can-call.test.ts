import { describe, expect, test } from "bun:test";
import { readFileSync, readdirSync } from "node:fs";
import { join } from "node:path";

// A page and its menu link ask canCall of the API operation that they call,
// so the permission comes from the operations of the catalog. A
// permission typed by hand drifts: the Admin menu once showed Admission to
// each role with audit:read, and the page refused the auditor (#1232).
describe("the UI checks the permissions of the operations it calls", () => {
  const src = join(import.meta.dir, "..");
  const files = readdirSync(src, { recursive: true, encoding: "utf8" }).filter(
    (f) =>
      /\.tsx?$/.test(f) &&
      !/\.(gen|test)\.tsx?$/.test(f) &&
      !f.startsWith("sdk/"),
  );

  test("no file checks a permission that it typed", () => {
    const typed = files.filter((f) =>
      /\bcan\(\s*["'`]/.test(readFileSync(join(src, f), "utf8")),
    );
    expect(typed).toEqual([]);
  });
});
