import { Picker } from "@/components/ui/popover";
import { useSchedules } from "@/hooks/use-schedules";
import { Calendar } from "lucide-react";

interface SchedulePickerProps {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
}

/** SchedulePicker picks a cron entry by name, once for all the agents that have it. */
export function SchedulePicker({
  id,
  label,
  value,
  onChange,
}: SchedulePickerProps) {
  const { entries, loading } = useSchedules();
  const names = [
    ...new Set(entries.map((e) => e.name).filter(Boolean)),
  ] as string[];

  return (
    <Picker
      id={id}
      label={label}
      value={value}
      onChange={onChange}
      loading={loading}
      placeholder="Select a cron entry"
      options={names.map((name) => ({
        value: name,
        icon: <Calendar className="h-3 w-3 shrink-0 text-text-muted" />,
      }))}
    />
  );
}
