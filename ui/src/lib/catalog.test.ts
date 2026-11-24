import { describe, expect, test } from "bun:test";
import {
  fieldValue,
  hasValue,
  methodInput,
  methodLabel,
  methodReady,
  methodRequest,
  type CatalogMethod,
} from "@/lib/catalog";

const method: CatalogMethod = {
  name: "update",
  kind: "write",
  doc: "Set the servers",
  http_method: "PUT",
  path: "/api/node/{hostname}/dns/{interface_name}",
  permission: "dns:write",
  platforms: ["debian"],
  runs_on: ["debian"],
  input: "UpdateInput { interface_name*, servers, override }",
  output: "Result { changed }",
  fields: [
    {
      name: "interface_name",
      in: "path",
      kind: "string",
      required: true,
      doc: "The interface",
    },
    { name: "servers", in: "body", kind: "list", required: false, doc: "" },
    { name: "override", in: "body", kind: "bool", required: false, doc: "" },
  ],
  columns: [],
  calls: [],
  examples: [],
};

describe("a method's request", () => {
  test("puts each field in its place, and keeps @ in a path", () => {
    const { url, init } = methodRequest(method, "web-01", {
      interface_name: "@fact.interface.primary",
      servers: ["1.1.1.1", "8.8.8.8"],
    });
    expect(url).toBe("/api/node/web-01/dns/@fact.interface.primary");
    expect(init.method).toBe("PUT");
    expect(JSON.parse(String(init.body))).toEqual({
      servers: ["1.1.1.1", "8.8.8.8"],
    });
  });

  test("a read puts its fields in the query and has no body", () => {
    const read: CatalogMethod = {
      ...method,
      kind: "read",
      http_method: "GET",
      path: "/api/node/{hostname}/package",
      fields: [
        { name: "names", in: "query", kind: "list", required: false, doc: "" },
      ],
    };
    const { url, init } = methodRequest(read, "_all", { names: "a, b" });
    expect(url).toBe("/api/node/_all/package?names=a&names=b");
    expect(init.body).toBeUndefined();
  });

  test("a flag goes once it is set, so an untouched one keeps the default", () => {
    expect(methodInput(method, { interface_name: "eth0" })).toEqual({
      interface_name: "eth0",
    });
    expect(
      methodInput(method, { interface_name: "eth0", override: false }),
    ).toEqual({ interface_name: "eth0", override: false });
  });
});

describe("a field's value", () => {
  test("is the kind that the API takes", () => {
    const f = method.fields[1];
    expect(fieldValue(f, "a, b,")).toEqual(["a", "b"]);
    expect(fieldValue(f, [" a ", ""])).toEqual(["a"]);
    expect(fieldValue({ ...f, kind: "number" }, "3")).toBe(3);
    expect(fieldValue({ ...f, kind: "bool" }, "true")).toBe(true);
    expect(fieldValue({ ...f, kind: "object" }, '{"a":1}')).toEqual({ a: 1 });
  });

  test("an empty list or text has no value", () => {
    expect(hasValue([])).toBe(false);
    expect(hasValue(" ")).toBe(false);
    expect(hasValue(["a"])).toBe(true);
  });
});

describe("methodReady", () => {
  test("is true when each required field has a value", () => {
    expect(methodReady(method, {})).toBe(false);
    expect(methodReady(method, { interface_name: "eth0" })).toBe(true);
  });
});

test("methodLabel is the name in words", () => {
  expect(methodLabel("update")).toBe("Update");
});
