import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";
import type { BlockStatus } from "@/hooks/use-stack";

interface DockerBlockProps {
  data: Record<string, unknown>;
  onChange: (data: Record<string, unknown>) => void;
  onStatusChange: (status: BlockStatus) => void;
}

export function DockerBlock({
  data,
  onChange,
  onStatusChange,
}: DockerBlockProps) {
  const image = (data.image as string) || "";
  const name = (data.name as string) || "";
  const ports = (data.ports as string) || "";
  const volumes = (data.volumes as string) || "";
  const env = (data.env as string) || "";
  const hostname = (data.hostname as string) || "";
  const dns = (data.dns as string) || "";
  const prevStatus = useRef<BlockStatus | null>(null);

  useEffect(() => {
    const next: BlockStatus = image.trim() !== "" ? "ready" : "pending";
    if (next !== prevStatus.current) {
      prevStatus.current = next;
      onStatusChange(next);
    }
  }, [image, onStatusChange]);

  const update = (field: string, value: string) => {
    onChange({ ...data, [field]: value });
  };

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-3">
        <Input
          id="docker-image"
          label="Image"
          placeholder="nginx:latest"
          value={image}
          onChange={(e) => update("image", e.target.value)}
        />
        <Input
          id="docker-name"
          label="Container Name (optional)"
          placeholder="my-nginx"
          value={name}
          onChange={(e) => update("name", e.target.value)}
        />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <Input
          id="docker-ports"
          label="Ports (comma-separated)"
          placeholder="8080:80, 443:443"
          value={ports}
          onChange={(e) => update("ports", e.target.value)}
        />
        <Input
          id="docker-volumes"
          label="Volumes (comma-separated)"
          placeholder="/host/path:/container/path"
          value={volumes}
          onChange={(e) => update("volumes", e.target.value)}
        />
      </div>
      <Input
        id="docker-env"
        label="Environment (comma-separated KEY=VALUE)"
        placeholder="NODE_ENV=production, PORT=3000"
        value={env}
        onChange={(e) => update("env", e.target.value)}
      />
      <div className="grid grid-cols-2 gap-3">
        <Input
          id="docker-hostname"
          label="Hostname (optional)"
          placeholder="my-container"
          value={hostname}
          onChange={(e) => update("hostname", e.target.value)}
        />
        <Input
          id="docker-dns"
          label="DNS Servers (optional)"
          placeholder="1.1.1.1, 8.8.8.8"
          value={dns}
          onChange={(e) => update("dns", e.target.value)}
        />
      </div>
    </div>
  );
}
