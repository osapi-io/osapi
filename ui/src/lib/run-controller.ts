import {
  deleteFileByName,
  getFileStale,
  getFiles,
  postFile,
} from "@/sdk/gen/file-management-api-file-operations/file-management-api-file-operations";
import { type PlanStep } from "@/lib/plan";

/** Runs a call of the controller: a step of the object store. */
export async function runController(
  step: PlanStep,
): Promise<{ data: unknown }> {
  switch (step.run) {
    case "objectstore.list":
      return getFiles();
    case "objectstore.upload":
      return postFile({
        name: step.with.name as string,
        content_type: ((step.with.content_type as string) || "raw") as
          "raw" | "template",
        file: step.with._file as globalThis.File,
      });
    case "objectstore.delete":
      return deleteFileByName(step.with.name as string);
    default:
      return getFileStale();
  }
}
