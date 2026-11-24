import { describe, expect, mock, test } from "bun:test";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { SecretForm } from "@/components/domain/secret-form";

const fill = (key: string, value: string) => {
  fireEvent.change(screen.getByLabelText("Key"), { target: { value: key } });
  fireEvent.change(screen.getByLabelText("Value"), {
    target: { value },
  });
};

describe("SecretForm", () => {
  test("the value is a password field, and Write needs a key and a value", () => {
    render(<SecretForm onWrite={async () => true} />);
    const value = screen.getByLabelText("Value") as HTMLInputElement;
    expect(value.type).toBe("password");
    const write = screen.getByRole("button", { name: "Write" });
    expect((write as HTMLButtonElement).disabled).toBe(true);
    fill("registry-token", "s3cret");
    expect((write as HTMLButtonElement).disabled).toBe(false);
  });

  test("a write sends the key and the value, then empties both", async () => {
    const onWrite = mock(async () => true);
    render(<SecretForm onWrite={onWrite} />);
    fill("registry-token", "s3cret");
    fireEvent.click(screen.getByRole("button", { name: "Write" }));
    await waitFor(() =>
      expect(onWrite).toHaveBeenCalledWith("registry-token", "s3cret"),
    );
    await waitFor(() =>
      expect((screen.getByLabelText("Value") as HTMLInputElement).value).toBe(
        "",
      ),
    );
    expect((screen.getByLabelText("Key") as HTMLInputElement).value).toBe("");
  });

  test("a write that fails empties the value and keeps the key", async () => {
    render(<SecretForm initialKey="db" onWrite={async () => false} />);
    fireEvent.change(screen.getByLabelText("Value"), {
      target: { value: "s3cret" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Write" }));
    await waitFor(() =>
      expect((screen.getByLabelText("Value") as HTMLInputElement).value).toBe(
        "",
      ),
    );
    expect((screen.getByLabelText("Key") as HTMLInputElement).value).toBe("db");
  });
});
