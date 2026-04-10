import { useContext, useEffect } from "react";
import { CommandRegistryContext, type Command } from "@/lib/command-context";

export function useCommandRegistry() {
  const ctx = useContext(CommandRegistryContext);
  if (!ctx)
    throw new Error(
      "useCommandRegistry must be used within CommandRegistryProvider",
    );
  return ctx;
}

/**
 * Register commands for the lifetime of the calling component.
 * Commands are unregistered on unmount.
 */
export function useCommands(commands: Command[], deps: unknown[] = []) {
  const { register } = useCommandRegistry();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => register(commands), deps);
}
