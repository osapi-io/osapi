import { useJob } from "@/hooks/use-job";
import { Link } from "react-router-dom";
import { SectionLabel } from "@/components/ui/section-label";
import { Text } from "@/components/ui/text";
import { Card, CardContent } from "@/components/ui/card";
import { Spinner } from "@/components/ui/spinner";
import { JobView } from "@/components/domain/job-detail";

// LinkedJob is the job that /admin/jobs/:id names, such as the job of a
// step's result on the Plans page: its detail, above the list, whether
// or not the job is on the page of the list that shows.
export function LinkedJob({ id }: { id: string }) {
  const { data: job, error: failed } = useJob(id);

  return (
    <div className="mb-4">
      <SectionLabel>
        Job <span className="font-mono">{id}</span>
        <Link
          to="/admin/jobs"
          className="ml-3 text-xs text-text-muted transition-colors hover:text-text"
        >
          close
        </Link>
      </SectionLabel>
      <Card>
        <CardContent>
          {failed ? (
            <Text variant="muted">{failed}</Text>
          ) : job ? (
            <JobView job={job} />
          ) : (
            <Spinner size="inline" />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
