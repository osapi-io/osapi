import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "gen/api/osapi",
    },
    {
      type: "category",
      label: "Admission",
      link: {
        type: "doc",
        id: "gen/api/admission-api-admission",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-admissions",
          label: "List admission records",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/delete-machine",
          label: "Delete a machine",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "gen/api/get-admission",
          label: "Get a machine's admission record",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/admit-machine",
          label: "Admit a pending machine",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/reject-machine",
          label: "Reject a pending machine",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/revoke-machine",
          label: "Revoke an admitted machine",
          className: "api-method post",
        },
      ],
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
          id: "gen/api/get-agents",
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
        {
          type: "doc",
          id: "gen/api/get-agent-status",
          label: "Show the health of each agent",
          className: "api-method get",
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
          id: "gen/api/get-audit-log-by-id",
          label: "Get a single audit log entry",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-audit-export",
          label: "Export all audit log entries",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Controller",
      link: {
        type: "doc",
        id: "gen/api/controller-api-controller",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-controller-status",
          label: "Show the health of the controller",
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
          id: "gen/api/get-files",
          label: "List stored files",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-file",
          label: "Upload a file",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/delete-file-by-name",
          label: "Delete a file",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "gen/api/get-file-by-name",
          label: "Get file metadata",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-file-by-name",
          label: "Replace a file",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "gen/api/get-file-stale",
          label: "List stale file deployments",
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
          id: "gen/api/get-jobs",
          label: "List jobs",
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
          id: "gen/api/get-job-by-id",
          label: "Get job detail",
          className: "api-method get",
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
      label: "NATS",
      link: {
        type: "doc",
        id: "gen/api/nats-api-nats",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-nats-status",
          label: "Show the health of the NATS server",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Telemetry",
      link: {
        type: "doc",
        id: "gen/api/telemetry-api-telemetry",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-trace",
          label: "Get the spans of a trace",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Certificate",
      link: {
        type: "doc",
        id: "gen/api/certificate-management-api-certificate-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-certificate-ca",
          label: "List all CA certificates",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-certificate-ca",
          label: "Create a CA certificate",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/put-node-certificate-ca",
          label: "Update a CA certificate",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-certificate-ca",
          label: "Delete a CA certificate",
          className: "api-method delete",
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
          id: "gen/api/get-node-file-status",
          label: "Check deployment status of a file on the host",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Log",
      link: {
        type: "doc",
        id: "gen/api/log-management-api-log-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-log",
          label: "Get system log entries",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-node-log-source",
          label: "List log sources",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-node-log-unit",
          label: "Get log entries for a systemd unit",
          className: "api-method get",
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
      label: "Node/Network/Route",
      link: {
        type: "doc",
        id: "gen/api/network-management-api-route-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-network-route",
          label: "List network routes",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/get-node-network-route-by-interface",
          label: "Get routes for an interface",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-network-route",
          label: "Create routes for an interface",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/put-node-network-route",
          label: "Update routes for an interface",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-network-route",
          label: "Delete routes for an interface",
          className: "api-method delete",
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
      label: "Node/Schedule",
      link: {
        type: "doc",
        id: "gen/api/schedule-management-api-schedule-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-schedule",
          label: "List all scheduled entries",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-schedule",
          label: "Create a scheduled entry",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/get-node-schedule-by-name",
          label: "Get a scheduled entry",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-node-schedule",
          label: "Update a scheduled entry",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-schedule",
          label: "Delete a scheduled entry",
          className: "api-method delete",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Service",
      link: {
        type: "doc",
        id: "gen/api/service-management-api-service-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-service",
          label: "List all services",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-service",
          label: "Create a service unit file",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/get-node-service-by-name",
          label: "Get service details",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-node-service",
          label: "Update a service unit file",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-service",
          label: "Delete a service unit file",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "gen/api/post-node-service-start",
          label: "Start a service",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/post-node-service-stop",
          label: "Stop a service",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/post-node-service-restart",
          label: "Restart a service",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/post-node-service-enable",
          label: "Enable a service",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/post-node-service-disable",
          label: "Disable a service",
          className: "api-method post",
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
        {
          type: "doc",
          id: "gen/api/get-node-user-ssh-key",
          label: "List SSH authorized keys",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-user-ssh-key",
          label: "Add SSH authorized key",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-user-ssh-key",
          label: "Remove SSH authorized key",
          className: "api-method delete",
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
    {
      type: "category",
      label: "Node/Disk",
      link: {
        type: "doc",
        id: "gen/api/disk-api-disk-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-disk",
          label: "Get the disk usage of the host",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Hostname",
      link: {
        type: "doc",
        id: "gen/api/hostname-api-hostname-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-hostname",
          label: "Get the hostname of the host",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-node-hostname",
          label: "Set the hostname of the host",
          className: "api-method put",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Load",
      link: {
        type: "doc",
        id: "gen/api/load-api-load-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-load",
          label: "Get the load averages of the host",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Memory",
      link: {
        type: "doc",
        id: "gen/api/memory-api-memory-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-memory",
          label: "Get the memory usage of the host",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Dns",
      link: {
        type: "doc",
        id: "gen/api/dns-api-network-dns-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/delete-node-network-dns",
          label: "Drop the DNS resolvers osapi set on an interface",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "gen/api/get-node-network-dns-by-key",
          label: "Get the DNS resolvers of an interface",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-node-network-dns",
          label: "Set the DNS resolvers of an interface",
          className: "api-method put",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Interface",
      link: {
        type: "doc",
        id: "gen/api/interface-api-network-interface-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-network-interface",
          label: "List the network interfaces of the host",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-network-interface",
          label: "Drop the config osapi has of a network interface",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "gen/api/get-node-network-interface-by-key",
          label: "Get a network interface of the host",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-network-interface",
          label: "Configure a network interface",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/put-node-network-interface",
          label: "Change the config osapi has of a network interface",
          className: "api-method put",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Os",
      link: {
        type: "doc",
        id: "gen/api/os-api-os-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-os",
          label: "Get the operating system of the host",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Package",
      link: {
        type: "doc",
        id: "gen/api/package-api-package-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-package",
          label: "List the packages installed on the host",
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
          id: "gen/api/delete-node-package",
          label: "Remove a package",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "gen/api/get-node-package-by-key",
          label: "Get one installed package",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Status",
      link: {
        type: "doc",
        id: "gen/api/status-api-status-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-status",
          label: "Get the status summary of the host",
          className: "api-method get",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Sysctl",
      link: {
        type: "doc",
        id: "gen/api/sysctl-api-sysctl-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-sysctl",
          label: "List the kernel's parameters and which osapi manages",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/post-node-sysctl",
          label: "Manage a kernel parameter and apply it",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api/delete-node-sysctl",
          label: "Stop the management of a kernel parameter",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "gen/api/get-node-sysctl-by-key",
          label: "Get a kernel parameter's current value, and whether osapi manages it",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-node-sysctl",
          label: "Set a kernel parameter, manage it, and apply it",
          className: "api-method put",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Timezone",
      link: {
        type: "doc",
        id: "gen/api/timezone-api-timezone-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-timezone",
          label: "Get the timezone of the host",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-node-timezone",
          label: "Set the timezone of the host",
          className: "api-method put",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Updatepackages",
      link: {
        type: "doc",
        id: "gen/api/updatepackages-api-updatepackages-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-updatepackages",
          label: "List the packages the index has newer versions of",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api/put-node-updatepackages",
          label: "Refresh the package index",
          className: "api-method put",
        },
      ],
    },
    {
      type: "category",
      label: "Node/Uptime",
      link: {
        type: "doc",
        id: "gen/api/uptime-api-uptime-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api/get-node-uptime",
          label: "Get the uptime of the host",
          className: "api-method get",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
