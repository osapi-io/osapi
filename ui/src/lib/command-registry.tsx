import { useState, useCallback, type ReactNode } from "react";
import { CommandRegistryContext, type Command } from "@/lib/command-context";

export type { Command };

export function CommandRegistryProvider({ children }: { children: ReactNode }) {
  const [commandSets, setCommandSets] = useState<Map<symbol, Command[]>>(
    new Map(),
  );

  const register = useCallback((cmds: Command[]) => {
    const key = Symbol();
    setCommandSets((prev) => new Map(prev).set(key, cmds));
    return () => {
      setCommandSets((prev) => {
        const next = new Map(prev);
        next.delete(key);
        return next;
      });
    };
  }, []);

  const commands = Array.from(commandSets.values()).flat();

  return (
    <CommandRegistryContext.Provider value={{ commands, register }}>
      {children}
    </CommandRegistryContext.Provider>
  );
}
