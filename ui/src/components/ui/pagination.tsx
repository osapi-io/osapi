import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Text } from "@/components/ui/text";

interface PaginationProps {
  offset: number;
  limit: number;
  total: number;
  onChange: (offset: number) => void;
}

/** Previous and Next under a list, with the page that shows. */
export function Pagination({
  offset,
  limit,
  total,
  onChange,
}: PaginationProps) {
  const pages = Math.max(1, Math.ceil(total / limit));
  if (pages <= 1) return null;
  const page = Math.floor(offset / limit) + 1;
  return (
    <div className="flex items-center justify-between">
      <Text variant="muted">
        Page {page} of {pages}
      </Text>
      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="sm"
          disabled={offset === 0}
          onClick={() => onChange(Math.max(0, offset - limit))}
        >
          <ChevronLeft className="h-4 w-4" />
          Previous
        </Button>
        <Button
          variant="ghost"
          size="sm"
          disabled={offset + limit >= total}
          onClick={() => onChange(offset + limit)}
        >
          Next
          <ChevronRight className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
