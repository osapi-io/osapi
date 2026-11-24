import { useState } from "react";

export function useListSearch() {
  const [search, setSearch] = useState("");
  return { search, setSearch };
}
