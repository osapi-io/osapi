import { Text } from "@/components/ui/text";
import { Input } from "@/components/ui/input";

export function FilterInput({
  label,
  value,
  placeholder,
  className = "w-48",
  onChange,
}: {
  label: string;
  value: string;
  placeholder: string;
  className?: string;
  onChange: (value: string) => void;
}) {
  return (
    <div className={className}>
      <Text variant="label" as="label" className="mb-1 block">
        {label}
      </Text>
      <Input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
      />
    </div>
  );
}
