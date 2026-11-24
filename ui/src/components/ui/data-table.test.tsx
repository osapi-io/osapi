import { describe, expect, mock, test } from "bun:test";
import { fireEvent, render, screen } from "@testing-library/react";
import { DataTable } from "@/components/ui/data-table";

type Row = { id: string; name: string };
const columns = [{ header: "Name", cell: (r: Row) => r.name }];

describe("DataTable", () => {
  test("a row for each item, under its headers", () => {
    render(
      <DataTable
        columns={columns}
        rows={[
          { id: "1", name: "web-01" },
          { id: "2", name: "web-02" },
        ]}
        getRowKey={(r) => r.id}
      />,
    );
    expect(screen.getByText("Name")).toBeTruthy();
    expect(screen.getByText("web-01")).toBeTruthy();
    expect(screen.getByText("web-02")).toBeTruthy();
  });

  test("an empty table says so, as the CLI does", () => {
    render(
      <DataTable columns={columns} rows={[]} getRowKey={(r: Row) => r.id} />,
    );
    expect(screen.getByText("None.")).toBeTruthy();
  });

  test("a click on a row gives the row", () => {
    const onRowClick = mock(() => {});
    render(
      <DataTable
        columns={columns}
        rows={[{ id: "1", name: "web-01" }]}
        getRowKey={(r) => r.id}
        onRowClick={onRowClick}
      />,
    );
    fireEvent.click(screen.getByText("web-01"));
    expect(onRowClick).toHaveBeenCalledWith({ id: "1", name: "web-01" });
  });
});
