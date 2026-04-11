import { createContext } from "react";

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

export interface CommandRegistry {
  commands: Command[];
  register: (commands: Command[]) => () => void;
}

export const CommandRegistryContext = createContext<CommandRegistry | null>(
  null,
);
