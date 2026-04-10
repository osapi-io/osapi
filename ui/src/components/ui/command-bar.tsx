import { useState, useEffect, useRef, useCallback } from "react";
import { cn } from "@/lib/cn";
import { useCommandRegistry } from "@/lib/command-registry";
import { features } from "@/lib/features";
import { Kbd } from "@/components/ui/kbd";

export function CommandBar() {
  const [open, setOpen] = useState(false);
  const [input, setInput] = useState("");
  const [activeIndex, setActiveIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const { commands } = useCommandRegistry();

  const filtered = input
    ? commands
        .filter(
          (c) =>
            c.name.toLowerCase().includes(input.toLowerCase()) ||
            c.description.toLowerCase().includes(input.toLowerCase()) ||
            (c.category ?? "").toLowerCase().includes(input.toLowerCase()),
        )
        .sort((a, b) => {
          const q = input.toLowerCase();
          const aName = a.name.toLowerCase();
          const bName = b.name.toLowerCase();
          // Exact start of name wins
          const aStarts = aName.startsWith(q) ? 0 : 1;
          const bStarts = bName.startsWith(q) ? 0 : 1;
          if (aStarts !== bStarts) return aStarts - bStarts;
          // Word boundary match beats substring
          const aWord = aName.split(" ").some((w) => w.startsWith(q)) ? 0 : 1;
          const bWord = bName.split(" ").some((w) => w.startsWith(q)) ? 0 : 1;
          if (aWord !== bWord) return aWord - bWord;
          // Name match beats description-only match
          const aInName = aName.includes(q) ? 0 : 1;
          const bInName = bName.includes(q) ? 0 : 1;
          return aInName - bInName;
        })
    : commands;

  const close = useCallback(() => {
    setOpen(false);
    setInput("");
    setActiveIndex(0);
  }, []);

  const execute = useCallback(
    (idx: number) => {
      const cmd = filtered[idx];
      if (!cmd) return;
      close();
      cmd.action();
    },
    [filtered, close],
  );

  // Open on : key (when not in an input) — vim command mode
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement).tagName;
      const inInput = tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT";

      if (e.key === ":" && !inInput && !open) {
        e.preventDefault();
        setOpen(true);
        setTimeout(() => inputRef.current?.focus(), 0);
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [open]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Escape") {
      e.preventDefault();
      close();
    } else if (e.key === "ArrowDown" || (e.ctrlKey && e.key === "j")) {
      e.preventDefault();
      const next = activeIndex < filtered.length - 1 ? activeIndex + 1 : 0;
      setActiveIndex(next);
      const buttons = listRef.current?.querySelectorAll("button");
      buttons?.[next]?.scrollIntoView({ block: "nearest" });
    } else if (e.key === "ArrowUp" || (e.ctrlKey && e.key === "k")) {
      e.preventDefault();
      const next = activeIndex > 0 ? activeIndex - 1 : filtered.length - 1;
      setActiveIndex(next);
      const buttons = listRef.current?.querySelectorAll("button");
      buttons?.[next]?.scrollIntoView({ block: "nearest" });
    } else if (e.key === "Enter" && filtered.length > 0) {
      e.preventDefault();
      execute(activeIndex);
    }
  };

  if (!open || !features.keyboard) return null;

  // Group by category
  const grouped = new Map<string, typeof filtered>();
  for (const cmd of filtered) {
    const cat = cmd.category ?? "";
    if (!grouped.has(cat)) grouped.set(cat, []);
    grouped.get(cat)!.push(cmd);
  }

  let flatIdx = 0;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh]">
      <div className="absolute inset-0 bg-black/60" onClick={close} />
      <div className="relative w-full max-w-md rounded-lg border border-border/60 bg-card shadow-2xl">
        <div className="flex items-center border-b border-border/30 px-4">
          <span className="font-mono text-sm text-primary">:</span>
          <input
            ref={inputRef}
            value={input}
            onChange={(e) => {
              setInput(e.target.value);
              setActiveIndex(0);
            }}
            onKeyDown={handleKeyDown}
            placeholder="Type a command..."
            autoComplete="off"
            className="flex-1 bg-transparent px-2 py-3 font-mono text-sm text-text placeholder:text-text-muted focus:outline-none"
          />
          <Kbd>esc</Kbd>
        </div>
        <div ref={listRef} className="max-h-72 overflow-auto py-1">
          {filtered.length === 0 ? (
            <p className="px-4 py-3 font-mono text-xs text-text-muted">
              No matching commands
            </p>
          ) : (
            Array.from(grouped.entries()).map(([category, cmds]) => {
              const items = cmds.map((cmd) => {
                const idx = flatIdx++;
                return (
                  <button
                    key={cmd.id}
                    onClick={() => execute(idx)}
                    className={cn(
                      "flex w-full items-center gap-3 px-4 py-2 text-left transition-colors",
                      idx === activeIndex
                        ? "bg-primary/20 text-primary border-l-2 border-primary"
                        : "text-text hover:bg-white/5 border-l-2 border-transparent",
                    )}
                  >
                    <span className="font-mono text-xs text-primary/70">:</span>
                    <span className="font-mono text-sm">{cmd.name}</span>
                    <span className="ml-auto text-xs text-text-muted">
                      {cmd.description}
                    </span>
                  </button>
                );
              });

              return (
                <div key={category || "__none"}>
                  {category && (
                    <div className="px-4 pb-1 pt-2">
                      <span className="font-mono text-[10px] uppercase tracking-wider text-text-muted/60">
                        {category}
                      </span>
                    </div>
                  )}
                  {items}
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
}
