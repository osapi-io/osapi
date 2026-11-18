import { getNodeContainerDocker } from "@/sdk/gen/docker-management-api-docker-operations/docker-management-api-docker-operations";
import type {
  DockerListCollectionResponse,
  DockerSummary,
} from "@/sdk/gen/schemas";
import { useLoad } from "@/hooks/use-load";

/** A container, with the host it runs on. */
export interface HostContainer extends DockerSummary {
  hostname: string;
}

const load = async () => {
  const r = await getNodeContainerDocker("_all");
  if (r.status !== 200) return undefined;
  return (r.data as DockerListCollectionResponse).results.flatMap((host) =>
    (host.containers ?? []).map((c) => ({ ...c, hostname: host.hostname })),
  );
};

/** The containers on every agent. */
export function useContainers() {
  const { value, loading } = useLoad<HostContainer[]>(load, []);
  return { containers: value, loading };
}
