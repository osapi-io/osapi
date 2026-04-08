import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import { FactInput } from "@/components/ui/fact-input";
import { Text } from "@/components/ui/text";
import type { BlockStatus } from "@/hooks/use-stack";

interface InterfaceBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function InterfaceBlock({
  data,
  onChange,
  onStatusChange,
}: InterfaceBlockProps) {
  const name = (data.name as string) || "";
  const addresses = (data.addresses as string) || "";
  const gateway4 = (data.gateway4 as string) || "";
  const gateway6 = (data.gateway6 as string) || "";
  const mtu = (data.mtu as string) || "";
  const macAddress = (data.mac_address as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus = name.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [name, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  const updateBool = (field: string, value: boolean) => {
    onChange({ ...data, [field]: value });
  };

  const dhcp4 = data.dhcp4 === true;
  const dhcp6 = data.dhcp6 === true;
  const wakeonlan = data.wakeonlan === true;

  return (
    <div className="space-y-3">
      <FactInput
        id="interface-name"
        label="Interface Name"
        placeholder="eth0 or @fact.interface.primary"
        value={name}
        onChange={(v) => update("name", v)}
      />
      <div className="grid grid-cols-2 gap-3">
        <Input
          id="interface-addresses"
          label="Addresses (comma-separated)"
          placeholder="192.168.1.10/24, fd00::1/64"
          value={addresses}
          onChange={(e) => update("addresses", e.target.value)}
        />
        <Input
          id="interface-gateway4"
          label="IPv4 Gateway"
          placeholder="192.168.1.1"
          value={gateway4}
          onChange={(e) => update("gateway4", e.target.value)}
        />
        <Input
          id="interface-gateway6"
          label="IPv6 Gateway"
          placeholder="fd00::1"
          value={gateway6}
          onChange={(e) => update("gateway6", e.target.value)}
        />
        <Input
          id="interface-mtu"
          label="MTU"
          placeholder="1500"
          value={mtu}
          onChange={(e) => update("mtu", e.target.value)}
        />
        <Input
          id="interface-mac"
          label="MAC Address"
          placeholder="aa:bb:cc:dd:ee:ff"
          value={macAddress}
          onChange={(e) => update("mac_address", e.target.value)}
        />
      </div>
      <div className="flex items-center gap-6">
        <label className="flex cursor-pointer items-center gap-2">
          <input
            type="checkbox"
            checked={dhcp4}
            onChange={(e) => updateBool("dhcp4", e.target.checked)}
            className="accent-primary"
          />
          <Text variant="muted">DHCP4</Text>
        </label>
        <label className="flex cursor-pointer items-center gap-2">
          <input
            type="checkbox"
            checked={dhcp6}
            onChange={(e) => updateBool("dhcp6", e.target.checked)}
            className="accent-primary"
          />
          <Text variant="muted">DHCP6</Text>
        </label>
        <label className="flex cursor-pointer items-center gap-2">
          <input
            type="checkbox"
            checked={wakeonlan}
            onChange={(e) => updateBool("wakeonlan", e.target.checked)}
            className="accent-primary"
          />
          <Text variant="muted">Wake-on-LAN</Text>
        </label>
      </div>
    </div>
  );
}
