import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "gen/api/agent-management-api",
    },
    {
      type: "category",
      label: "Agent",
      link: {
        type: "doc",
        id: "gen/api/agent-management-api-agent-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/list-active-agents",
          label: "List active agents",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-agent-details",
          label: "Get agent details",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/drain-agent",
          label: "Drain an agent",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/undrain-agent",
          label: "Undrain an agent",
          className: "api-method post",
        },
      ],
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
      label: "Node/Docker",
      link: {
        type: "doc",
        id: "gen/api/docker-management-api-docker-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/post-node-container-docker",
          label: "Create a container",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/get-node-container-docker",
          label: "List containers",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-node-container-docker-by-id",
          label: "Inspect a container",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-container-docker-by-id",
          label: "Remove a container",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "gen/api/post-node-container-docker-start",
          label: "Start a container",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/post-node-container-docker-stop",
          label: "Stop a container",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Docker/Exec",
      link: {
        type: "doc",
        id: "gen/api/docker-management-api-docker-exec",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/post-node-container-docker-exec",
          label: "Execute a command in a container",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Docker/Image",
      link: {
        type: "doc",
        id: "gen/api/docker-management-api-docker-image",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/delete-node-container-docker-image",
          label: "Remove a container image",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "gen/api/post-node-container-docker-pull",
          label: "Pull a container image",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "File",
      link: {
        type: "doc",
        id: "gen/api/file-management-api-file-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/post-file",
          label: "Upload a file",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/get-files",
          label: "List stored files",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-file-by-name",
          label: "Get file metadata",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/delete-file-by-name",
          label: "Delete a file",
          className: "api-method delete",
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
          id: "gen/api/list-jobs",
          label: "List jobs",
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
      label: "Node",
      link: {
        type: "doc",
        id: "gen/api/node-management-api-node-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-hostname",
          label: "Retrieve node hostname",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-node-disk",
          label: "Retrieve disk usage",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-node-memory",
          label: "Retrieve memory stats",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-node-load",
          label: "Retrieve load averages",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-node-os",
          label: "Retrieve OS info",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-node-uptime",
          label: "Retrieve uptime",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-file-deploy",
          label: "Deploy a file from Object Store to the host",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/post-node-file-undeploy",
          label: "Remove a deployed file from the host filesystem",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/post-node-file-status",
          label: "Check deployment status of a file on the host",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Status",
      link: {
        type: "doc",
        id: "gen/api/node-management-api-node-status",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-status",
          label: "Retrieve node status",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Network",
      link: {
        type: "doc",
        id: "gen/api/node-management-api-network-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/post-node-network-ping",
          label: "Ping a remote server",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Network/DNS",
      link: {
        type: "doc",
        id: "gen/api/node-management-api-dns-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-network-dns-by-interface",
          label: "List DNS servers",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-node-network-dns",
          label: "Update DNS servers",
          className: "api-method put",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Command",
      link: {
        type: "doc",
        id: "gen/api/node-management-api-command-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/post-node-command-exec",
          label: "Execute a command",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/post-node-command-shell",
          label: "Execute a shell command",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Schedule/Cron",
      link: {
        type: "doc",
        id: "gen/api/schedule-management-api-cron-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-schedule-cron",
          label: "List all cron entries",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-schedule-cron",
          label: "Create a cron entry",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/get-node-schedule-cron-by-name",
          label: "Get a cron entry",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-node-schedule-cron",
          label: "Update a cron entry",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-schedule-cron",
          label: "Delete a cron entry",
          className: "api-method delete",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
