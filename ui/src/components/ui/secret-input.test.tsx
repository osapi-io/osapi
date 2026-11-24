import { describe, expect, mock, test } from "bun:test";
import { fireEvent, render, screen } from "@testing-library/react";
import { SecretInput } from "@/components/ui/secret-input";
import { isSecretRef, secretRef } from "@/lib/secret-ref";

const refs = [
  { ref: "secret://default/api-token", vault: "default", key: "api-token" },
  { ref: "secret://prod/db-password", vault: "prod", key: "db-password" },
];

describe("SecretInput", () => {
  test("a value is a password field, and a reference is text", () => {
    const { rerender } = render(
      <SecretInput
        id="s"
        label="Password"
        value="hunter2"
        onChange={() => {}}
        refs={refs}
        picker
      />,
    );
    const input = screen.getByLabelText("Password") as HTMLInputElement;
    expect(input.type).toBe("password");
    rerender(
      <SecretInput
        id="s"
        label="Password"
        value="secret://default/api-token"
        onChange={() => {}}
        refs={refs}
        picker
      />,
    );
    expect(input.type).toBe("text");
  });

  test("a value that starts as the scheme does stays hidden", () => {
    render(
      <SecretInput
        id="s"
        label="Password"
        value="secret1"
        onChange={() => {}}
        refs={refs}
        picker
      />,
    );
    const input = screen.getByLabelText("Password") as HTMLInputElement;
    expect(input.type).toBe("password");
  });

  test("the picker lists each reference by vault, and a click inserts it", () => {
    const onChange = mock((_v: string) => {});
    render(
      <SecretInput
        id="s"
        label="Password"
        value=""
        onChange={onChange}
        refs={refs}
        picker
      />,
    );
    fireEvent.click(screen.getByTitle("Insert secret reference"));
    expect(screen.getByText("default")).toBeTruthy();
    expect(screen.getByText("prod")).toBeTruthy();
    fireEvent.click(screen.getByText("secret://prod/db-password"));
    expect(onChange).toHaveBeenCalledWith("secret://prod/db-password");
  });

  test("typing a reference filters the picker", () => {
    render(
      <SecretInput
        id="s"
        label="Password"
        value="secret://prod"
        onChange={() => {}}
        refs={refs}
        picker
      />,
    );
    fireEvent.click(screen.getByTitle("Insert secret reference"));
    expect(screen.queryByText("secret://default/api-token")).toBeNull();
    expect(screen.getByText("secret://prod/db-password")).toBeTruthy();
  });

  test("with no picker, the field has no button", () => {
    render(
      <SecretInput
        id="s"
        label="Password"
        value=""
        onChange={() => {}}
        refs={[]}
        picker={false}
      />,
    );
    expect(screen.queryByTitle("Insert secret reference")).toBeNull();
  });

  test("isSecretRef is a reference or the start of one", () => {
    expect(isSecretRef("secret://default/a")).toBe(true);
    expect(isSecretRef("secr")).toBe(true);
    expect(isSecretRef("hunter2")).toBe(false);
    expect(isSecretRef("")).toBe(false);
    expect(secretRef("default", "a")).toBe("secret://default/a");
  });
});
