import { cn } from "@/lib/cn";
import { Card, CardContent } from "@/components/ui/card";
import { Fragment, type ReactNode } from "react";

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
  /** What an empty table says, in one row; the CLI says "None.". */
  empty?: string;
  /** Called when a row is clicked, such as to expand it. */
  onRowClick?: (row: T) => void;
  /** What shows in a full-width row under a row, when it is not undefined. */
  expanded?: (row: T) => ReactNode | undefined;
  /** The group of a row: a row of its own heads each group, in order. */
  group?: (row: T) => string;
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
  empty = "None.",
  onRowClick,
  expanded,
  group,
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
                    "font-medium uppercase tracking-wide",
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
            {rows.length === 0 && (
              <tr>
                <td
                  colSpan={columns.length}
                  className={cn(hPad, cVPad, "text-text-muted")}
                >
                  {empty}
                </td>
              </tr>
            )}
            {rows.map((row, rowIdx) => {
              const heading = group?.(row);
              const startsGroup =
                heading !== undefined &&
                (rowIdx === 0 || group?.(rows[rowIdx - 1]) !== heading);
              const detail = expanded?.(row);

              return (
                <Fragment key={getRowKey(row, rowIdx)}>
                  {startsGroup && (
                    <tr>
                      <td
                        colSpan={columns.length}
                        className={cn(
                          hPad,
                          cVPad,
                          "border-t border-border/30 font-semibold text-text-muted",
                        )}
                      >
                        {heading}
                      </td>
                    </tr>
                  )}
                  <tr
                    className={cn(
                      "border-b border-border/20 last:border-0",
                      onRowClick &&
                        "cursor-pointer transition-colors hover:bg-white/[0.02]",
                    )}
                    onClick={onRowClick && (() => onRowClick(row))}
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
                  {detail !== undefined && (
                    <tr className="border-b border-border/20">
                      <td colSpan={columns.length} className="px-6 py-3">
                        {detail}
                      </td>
                    </tr>
                  )}
                </Fragment>
              );
            })}
          </tbody>
        </table>
      </CardContent>
    </Card>
  );
}
