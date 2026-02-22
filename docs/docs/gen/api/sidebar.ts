import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "gen/api/audit-log-api",
    },
    {
      type: "category",
      label: "Audit",
      link: {
        type: "doc",
        id: "gen/api/audit-log-api-audit",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-audit-logs",
          label: "List audit log entries",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-audit-export",
          label: "Export all audit log entries",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-audit-log-by-id",
          label: "Get a single audit log entry",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Info",
      link: {
        type: "doc",
        id: "gen/api/osapi-a-crud-api-for-managing-linux-systems-info",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/retrieve-the-software-version",
          label: "Retrieve the software version",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Health",
      link: {
        type: "doc",
        id: "gen/api/health-check-api-health",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-health",
          label: "Liveness probe",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-health-ready",
          label: "Readiness probe",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-health-status",
          label: "System status and component health",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Job",
      link: {
        type: "doc",
        id: "gen/api/job-management-api-job-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/create-a-new-job",
          label: "Create a new job",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/list-jobs",
          label: "List jobs",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-queue-statistics",
          label: "Get queue statistics",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/list-active-workers",
          label: "List active workers",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-job-by-id",
          label: "Get job detail",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/delete-job-by-id",
          label: "Delete a job",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "gen/api/retry-job-by-id",
          label: "Retry a job",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Network",
      link: {
        type: "doc",
        id: "gen/api/network-management-api-network-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/ping-a-remote-server",
          label: "Ping a remote server",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Network/DNS",
      link: {
        type: "doc",
        id: "gen/api/network-management-api-dns-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-network-dns-by-interface",
          label: "List DNS servers",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-network-dns",
          label: "Update DNS servers",
          className: "api-method put",
        },
      ],
    },
    {
      type: "category",
      label: "System",
      link: {
        type: "doc",
        id: "gen/api/system-management-api-system-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/retrieve-system-hostname",
          label: "Retrieve system hostname",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "System/Status",
      link: {
        type: "doc",
        id: "gen/api/system-management-api-system-status",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/retrieve-system-status",
          label: "Retrieve system status",
          className: "api-method get",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
