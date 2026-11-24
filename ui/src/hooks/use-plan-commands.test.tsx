import { describe, expect, mock, test } from "bun:test";
import { render, waitFor } from "@testing-library/react";
import { CommandRegistryProvider } from "@/lib/command-registry";
import { usePlanCommands } from "@/hooks/use-plan-commands";

const defs: [] = [];
const allowed = () => true;

describe("usePlanCommands", () => {
  test("new callbacks on each render do not register again and again", async () => {
    const renders = mock(() => {});
    function Page() {
      renders();
      usePlanCommands({
        canRun: false,
        running: false,
        run: () => {},
        clear: () => {},
        // A new function on each render, as the Plans page passed it.
        showYAML: () => {},
        defs,
        allowed,
        add: () => {},
      });
      return null;
    }
    render(
      <CommandRegistryProvider>
        <Page />
      </CommandRegistryProvider>,
    );
    await new Promise((r) => setTimeout(r, 50));
    await waitFor(() => expect(renders.mock.calls.length).toBeLessThan(5));
  });
});
