import { getNodeSchedule } from "@/sdk/gen/schedule-management-api-schedule-operations/schedule-management-api-schedule-operations";
import type {
  ScheduleCollectionResponse,
  ScheduleEntry,
} from "@/sdk/gen/schemas";
import { useLoad } from "@/hooks/use-load";

const load = async () => {
  const r = await getNodeSchedule("_all");
  return r.status === 200
    ? ((r.data as ScheduleCollectionResponse).results ?? [])
    : undefined;
};

/** The cron entries on every agent. */
export function useSchedules() {
  const { value, loading } = useLoad<ScheduleEntry[]>(load, []);
  return { entries: value, loading };
}
