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
      label: "Facts",
      link: {
        type: "doc",
        id: "gen/api/facts-api-facts",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-fact-keys",
          label: "List available fact keys",
          className: "api-method get",
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
      label: "Node/Command",
      link: {
        type: "doc",
        id: "gen/api/command-execution-api-command-operations",
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
      label: "Node/File",
      link: {
        type: "doc",
        id: "gen/api/node-file-operations-api-node-file-operations",
      },
      items: [
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
      label: "Node/Hostname",
      link: {
        type: "doc",
        id: "gen/api/hostname-management-api-hostname-operations",
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
          id: "gen/api/put-node-hostname",
          label: "Update node hostname",
          className: "api-method put",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Network",
      link: {
        type: "doc",
        id: "gen/api/network-management-api-network-operations",
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
        id: "gen/api/network-management-api-dns-operations",
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
      label: "Node/NTP",
      link: {
        type: "doc",
        id: "gen/api/ntp-management-api-ntp-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-ntp",
          label: "Get NTP status",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-ntp",
          label: "Create NTP configuration",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/put-node-ntp",
          label: "Update NTP configuration",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-ntp",
          label: "Delete NTP configuration",
          className: "api-method delete",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Package",
      link: {
        type: "doc",
        id: "gen/api/package-management-api-package-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-package",
          label: "List installed packages",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-package",
          label: "Install a package",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/get-node-package-by-name",
          label: "Get a package",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-package",
          label: "Remove a package",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "gen/api/post-node-package-update",
          label: "Update package sources",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/get-node-package-update",
          label: "List available updates",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Power",
      link: {
        type: "doc",
        id: "gen/api/power-management-api-power-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/post-node-power-reboot",
          label: "Reboot node",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/post-node-power-shutdown",
          label: "Shutdown node",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Process",
      link: {
        type: "doc",
        id: "gen/api/process-management-api-process-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-process",
          label: "List processes",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-node-process-by-pid",
          label: "Get process by PID",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-process-signal",
          label: "Send signal to process",
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
    {
      type: "category",
      label: "Node/Sysctl",
      link: {
        type: "doc",
        id: "gen/api/sysctl-management-api-sysctl-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-sysctl",
          label: "List all managed sysctl entries",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-sysctl",
          label: "Create a sysctl parameter",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/get-node-sysctl-by-key",
          label: "Get a sysctl entry",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-node-sysctl",
          label: "Update a sysctl parameter",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-sysctl",
          label: "Delete a managed sysctl entry",
          className: "api-method delete",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Timezone",
      link: {
        type: "doc",
        id: "gen/api/timezone-management-api-timezone-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-timezone",
          label: "Get system timezone",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-node-timezone",
          label: "Update system timezone",
          className: "api-method put",
        },
      ],
    },
    {
      type: "category",
      label: "Node/User",
      link: {
        type: "doc",
        id: "gen/api/user-and-group-management-api-user-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-user",
          label: "List all users",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-user",
          label: "Create a user",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/get-node-user-by-name",
          label: "Get a user",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-node-user",
          label: "Update a user",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-user",
          label: "Delete a user",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "gen/api/post-node-user-password",
          label: "Change user password",
          className: "api-method post",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Group",
      link: {
        type: "doc",
        id: "gen/api/user-and-group-management-api-group-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-group",
          label: "List all groups",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-group",
          label: "Create a group",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/get-node-group-by-name",
          label: "Get a group",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-node-group",
          label: "Update a group",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-group",
          label: "Delete a group",
          className: "api-method delete",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
