import { useEffect, useState } from "react";
import { getFiles } from "@/sdk/gen/file-management-api-file-operations/file-management-api-file-operations";
import type { FileInfo, FileListResponse } from "@/sdk/gen/schemas";

export type { FileInfo };

export function useObjects() {
  const [objects, setObjects] = useState<FileInfo[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;

    const fetch = async () => {
      try {
        const result = await getFiles();
        if (mounted && result.status === 200) {
          setObjects((result.data as FileListResponse).files ?? []);
        }
      } catch {
        // Silently fail — dropdown just stays empty
      }
      if (mounted) setLoading(false);
    };

    fetch();
    return () => {
      mounted = false;
    };
  }, []);

  const refresh = async () => {
    try {
      const result = await getFiles();
      if (result.status === 200) {
        setObjects((result.data as FileListResponse).files ?? []);
      }
    } catch {
      // ignore
    }
  };

  return { objects, loading, refresh };
}
