import { Picker } from "@/components/ui/popover";
import { useContainers } from "@/hooks/use-containers";
import { statusText } from "@/lib/status";
import { Text } from "@/components/ui/text";
import { Container } from "lucide-react";

interface ContainerPickerProps {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
}

/** ContainerPicker picks a container on any agent, by its name or its ID. */
export function ContainerPicker({
  id,
  label,
  value,
  onChange,
}: ContainerPickerProps) {
  const { containers, loading } = useContainers();

  return (
    <Picker
      id={id}
      label={label}
      value={value}
      onChange={onChange}
      loading={loading}
      placeholder="Select a container"
      panelClassName="max-h-60"
      options={containers.map((c) => ({
        value: c.name || c.id || "",
        key: `${c.hostname}-${c.id}`,
        label: c.name || c.id?.slice(0, 12),
        icon: <Container className="h-3 w-3 shrink-0 text-text-muted" />,
        meta: (
          <>
            {c.hostname}
            {c.state && (
              <Text variant={statusText(c.state)} className="ml-2 text-[10px]">
                {c.state}
              </Text>
            )}
          </>
        ),
      }))}
    />
  );
}
