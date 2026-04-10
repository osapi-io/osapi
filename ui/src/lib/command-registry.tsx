import {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  type ReactNode,
} from "react";

export interface Command {
  /** Unique command ID */
  id: string;
  /** Display name (shown after /) */
  name: string;
  /** Short description */
  description: string;
  /** Optional category for grouping */
  category?: string;
  /** The action to run */
  action: () => void;
}

interface CommandRegistry {
  commands: Command[];
  register: (commands: Command[]) => () => void;
}

const CommandRegistryContext = createContext<CommandRegistry | null>(null);

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
