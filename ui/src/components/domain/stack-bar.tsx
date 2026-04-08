import { useState, useRef, useEffect } from "react";
import { Layers, Plus, Search, Shield } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Text } from "@/components/ui/text";
import { cn } from "@/lib/cn";
import type { Stack } from "@/hooks/use-stacks";
import { SearchBox } from "@/components/ui/search-box";
import { IconButton } from "@/components/ui/icon-button";
import { ScrollButton } from "@/components/ui/scroll-button";

interface StackBarProps {
  stacks: Stack[];
  activeStackId: string | null;
  onLoad: (id: string) => void;
  onNew: () => void;
}

function permVariant(perm: string) {
  if (perm.endsWith(":write") || perm.endsWith(":execute")) {
    return "running" as const;
  }
  return "muted" as const;
}

function StackCard({
  stack,
  active,
  onClick,
}: {
  stack: Stack;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "group flex w-56 shrink-0 flex-col rounded-lg border px-3 py-2.5 text-left transition-all",
        active
          ? "border-primary/40 bg-primary/5 shadow-[0_0_12px_rgba(103,234,148,0.08)]"
          : "border-border/60 bg-card hover:border-primary/20 hover:bg-white/[0.02]",
      )}
    >
      <div className="flex items-center gap-2">
        <Layers
          className={cn(
            "h-3.5 w-3.5 shrink-0",
            active ? "text-primary" : "text-text-muted",
          )}
        />
        <Text size="sm" className="truncate font-medium">
          {stack.name}
        </Text>
      </div>
      <Text variant="muted" as="p" truncate className="mt-1">
        {stack.description}
      </Text>
      <div className="mt-1.5 flex items-center gap-2">
        <Text variant="muted">
          {stack.blocks.length} block{stack.blocks.length !== 1 ? "s" : ""}
        </Text>
        {stack.permissions.map((perm) => (
          <Badge key={perm} variant={permVariant(perm)}>
            <Shield className="h-3 w-3" />
            {perm}
          </Badge>
        ))}
      </div>
    </button>
  );
}

export function StackBar({
  stacks,
  activeStackId,
  onLoad,
  onNew,
}: StackBarProps) {
  const [search, setSearch] = useState("");
  const [showSearch, setShowSearch] = useState(false);
  const scrollRef = useRef<HTMLDivElement>(null);
  const [canScrollLeft, setCanScrollLeft] = useState(false);
  const [canScrollRight, setCanScrollRight] = useState(false);

  const filtered = search
    ? stacks.filter(
        (s) =>
          s.name.toLowerCase().includes(search.toLowerCase()) ||
          s.description.toLowerCase().includes(search.toLowerCase()),
      )
    : stacks;

  const checkScroll = () => {
    const el = scrollRef.current;
    if (!el) return;
    setCanScrollLeft(el.scrollLeft > 0);
    setCanScrollRight(el.scrollLeft + el.clientWidth < el.scrollWidth - 1);
  };

  useEffect(() => {
    checkScroll();
    const el = scrollRef.current;
    if (!el) return;
    el.addEventListener("scroll", checkScroll);
    const observer = new ResizeObserver(checkScroll);
    observer.observe(el);
    return () => {
      el.removeEventListener("scroll", checkScroll);
      observer.disconnect();
    };
  }, [filtered.length]);

  const scroll = (dir: "left" | "right") => {
    scrollRef.current?.scrollBy({
      left: dir === "left" ? -240 : 240,
      behavior: "smooth",
    });
  };

  return (
    <div className="mb-6 rounded-lg border border-border/60 bg-card/50">
      {/* Header */}
      <div className="flex items-center gap-3 border-b border-border/40 px-4 py-2">
        <Layers className="h-4 w-4 text-primary" />
        <span className="text-xs font-semibold uppercase tracking-wider text-text-muted">
          Stacks
        </span>
        <Text variant="muted">{stacks.length} saved</Text>

        <div className="ml-auto flex items-center gap-1">
          {showSearch ? (
            <SearchBox
              value={search}
              onChange={setSearch}
              onClose={() => {
                setSearch("");
                setShowSearch(false);
              }}
              placeholder="Search stacks..."
            />
          ) : (
            <IconButton
              icon={Search}
              variant="ghost"
              size="md"
              onClick={() => setShowSearch(true)}
            />
          )}
          <button
            onClick={onNew}
            className="flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-primary hover:bg-primary/10"
          >
            <Plus className="h-3.5 w-3.5" />
            New
          </button>
        </div>
      </div>

      {/* Scrollable cards */}
      <div className="flex items-center gap-1 px-2 py-3">
        <ScrollButton
          direction="left"
          disabled={!canScrollLeft}
          onClick={() => scroll("left")}
        />
        <div
          ref={scrollRef}
          className="flex min-w-0 flex-1 gap-2 overflow-x-auto px-2 scrollbar-none"
          style={{ scrollbarWidth: "none" }}
        >
          {filtered.length === 0 ? (
            <Text variant="muted" as="p" className="py-2">
              {search ? "No stacks match your search" : "No saved stacks"}
            </Text>
          ) : (
            filtered.map((stack) => (
              <StackCard
                key={stack.id}
                stack={stack}
                active={stack.id === activeStackId}
                onClick={() => onLoad(stack.id)}
              />
            ))
          )}
        </div>
        <ScrollButton
          direction="right"
          disabled={!canScrollRight}
          onClick={() => scroll("right")}
        />
      </div>
    </div>
  );
}
