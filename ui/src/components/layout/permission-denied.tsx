import { ContentArea } from "@/components/layout/content-area";
import { ErrorBanner } from "@/components/ui/error-banner";
import { PageHeader } from "@/components/ui/page-header";

interface PermissionDeniedProps {
  /** The page's title, as it shows when the token may read it. */
  title: string;
  /** What the page shows, such as "audit entries". */
  what: string;
}

/**
 * PermissionDenied is the page that a token without the page's permission
 * gets, in place of the page: its title, and that the token cannot read
 * what it shows.
 */
export function PermissionDenied({ title, what }: PermissionDeniedProps) {
  return (
    <ContentArea>
      <PageHeader title={title} subtitle="Insufficient permissions" />
      <ErrorBanner message={`You do not have permission to read ${what}.`} />
    </ContentArea>
  );
}
