import { describe, expect, test } from "bun:test";
import { jwtClaimStrings } from "@/lib/jwt";

const tokenWith = (payload: object) =>
  `e30.${btoa(JSON.stringify(payload)).replace(/=+$/, "")}.sig`;

describe("jwtClaimStrings", () => {
  test("the targets claim is the token's target scope", () => {
    const token = tokenWith({ roles: ["operator"], targets: ["group:web"] });
    expect(jwtClaimStrings(token, "targets")).toEqual(["group:web"]);
    expect(jwtClaimStrings(token, "roles")).toEqual(["operator"]);
  });

  test("a token without the claim has an empty list", () => {
    expect(jwtClaimStrings(tokenWith({ roles: ["admin"] }), "targets")).toEqual(
      [],
    );
  });

  test("a token that is not a JWT has an empty list", () => {
    expect(jwtClaimStrings("not-a-jwt", "roles")).toEqual([]);
  });
});
