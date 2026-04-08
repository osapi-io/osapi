import { cn } from "@/lib/cn";
import { Card, CardContent } from "@/components/ui/card";
import type { ReactNode } from "react";

interface ColumnDef<T> {
  header: string;
  align?: "left" | "right" | "center";
  cell: (row: T) => ReactNode;
  className?: string;
}

interface DataTableProps<T> {
  columns: ColumnDef<T>[];
  rows: T[];
  getRowKey: (row: T, index: number) => string | number;
  compact?: boolean;
  className?: string;
}

const alignClass = {
  left: "text-left",
  right: "text-right",
  center: "text-center",
};

export function DataTable<T>({
  columns,
  rows,
  getRowKey,
  compact = true,
  className,
}: DataTableProps<T>) {
  const hPad = compact ? "px-3" : "px-4";
  const hVPad = compact ? "py-2" : "py-2.5";
  const cVPad = compact ? "py-1.5" : "py-2";

  return (
    <Card className={className}>
      <CardContent className="p-0">
        <table className="w-full text-xs">
          <thead>
            <tr className="border-b border-border/40 text-left text-text-muted">
              {columns.map((col, i) => (
                <th
                  key={i}
                  className={cn(
                    hPad,
                    hVPad,
                    "font-medium",
                    alignClass[col.align ?? "left"],
                    col.className,
                  )}
                >
                  {col.header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, rowIdx) => (
              <tr
                key={getRowKey(row, rowIdx)}
                className="border-b border-border/20 last:border-0"
              >
                {columns.map((col, colIdx) => (
                  <td
                    key={colIdx}
                    className={cn(
                      hPad,
                      cVPad,
                      alignClass[col.align ?? "left"],
                      col.className,
                    )}
                  >
                    {col.cell(row)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </CardContent>
    </Card>
  );
}
