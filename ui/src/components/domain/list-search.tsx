import { SearchBox } from "@/components/ui/search-box";

export function ListSearchBar({
  search,
  setSearch,
  hasItems,
}: {
  search: string;
  setSearch: (v: string) => void;
  hasItems: boolean;
}) {
  if (!hasItems) return null;
  return (
    <div className="border-b border-border/40 px-4 py-2">
      <SearchBox
        value={search}
        onChange={setSearch}
        onClose={() => setSearch("")}
        placeholder="Filter..."
        autoFocus={false}
      />
    </div>
  );
}
