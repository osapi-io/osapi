import { Picker } from "@/components/ui/popover";
import { useTargets, type TargetOption } from "@/hooks/use-targets";
import { Globe, Monitor, Tag } from "lucide-react";

interface TargetPickerProps {
  value: string;
  onChange: (target: string) => void;
  label?: string;
  className?: string;
}

const groups: Record<
  TargetOption["group"],
  { heading: string; icon: typeof Globe }
> = {
  builtin: { heading: "Targets", icon: Globe },
  hostname: { heading: "Hosts", icon: Monitor },
  label: { heading: "Labels", icon: Tag },
};

/** TargetPicker picks the target of a step: _all, _any, a host or a label. */
export function TargetPicker({
  value,
  onChange,
  label = "Target",
  className,
}: TargetPickerProps) {
  const { options } = useTargets();
  const icon = (g: TargetOption["group"]) => {
    const Icon = groups[g].icon;
    return <Icon className="h-3 w-3 shrink-0 text-text-muted" />;
  };
  const selected = options.find((o) => o.value === value);

  return (
    <Picker
      label={label}
      value={value}
      onChange={onChange}
      className={className}
      placeholder="Select a target"
      panelClassName="max-h-56"
      display={
        selected ? (
          <>
            {icon(selected.group)}
            {selected.label}
          </>
        ) : undefined
      }
      options={options.map((o) => ({
        value: o.value,
        label: o.label,
        group: groups[o.group].heading,
        icon: icon(o.group),
      }))}
    />
  );
}
