import { getFiles } from "@/sdk/gen/file-management-api-file-operations/file-management-api-file-operations";
import type { FileInfo, FileListResponse } from "@/sdk/gen/schemas";
import { useLoad } from "@/hooks/use-load";

export type { FileInfo };

const load = async () => {
  const r = await getFiles();
  return r.status === 200
    ? ((r.data as FileListResponse).items ?? [])
    : undefined;
};

/** The objects in the file store. */
export function useObjects() {
  const { value, loading, refresh } = useLoad<FileInfo[]>(load, []);
  return { objects: value, loading, refresh };
}
