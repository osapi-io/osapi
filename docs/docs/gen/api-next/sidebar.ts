import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "gen/api-next/sysctl-api",
    },
    {
      type: "category",
      label: "Node/Sysctl",
      link: {
        type: "doc",
        id: "gen/api-next/sysctl-operations",
      },
      items: [
        {
          type: "doc",
          id: "gen/api-next/get-node-sysctl",
          label: "List the kernel's parameters and which osapi manages",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api-next/post-node-sysctl",
          label: "Manage a kernel parameter and apply it",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "gen/api-next/delete-node-sysctl",
          label: "Stop the management of a kernel parameter",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "gen/api-next/get-node-sysctl-by-key",
          label: "Get a kernel parameter's current value, and whether osapi manages it",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api-next/put-node-sysctl",
          label: "Set a kernel parameter, manage it, and apply it",
          className: "api-method put",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
