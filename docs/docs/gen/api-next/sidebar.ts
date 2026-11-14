import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "gen/api-next/osapi-behaviors",
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
          label: "List the kernel parameters osapi manages",
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
          id: "gen/api-next/get-node-sysctl-by-key",
          label: "Get a kernel parameter's current value",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "gen/api-next/put-node-sysctl",
          label: "Change a managed kernel parameter and apply it",
          className: "api-method put",
        },
        {
          type: "doc",
          id: "gen/api-next/delete-node-sysctl",
          label: "Stop managing a kernel parameter",
          className: "api-method delete",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
