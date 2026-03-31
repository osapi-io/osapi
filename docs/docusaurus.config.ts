// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

import type * as Preset from '@docusaurus/preset-classic';
import type { Config } from '@docusaurus/types';
import type * as Plugin from '@docusaurus/types/src/plugin';
import type * as OpenApiPlugin from 'docusaurus-plugin-openapi-docs';
import { themes as prismThemes } from 'prism-react-renderer';

const config: Config = {
  title: 'OSAPI',
  tagline: 'OSAPI is cool',
  favicon: 'img/favicon.ico',
  markdown: {
    mermaid: true,
    format: 'detect'
  },

  // Set the production url of your site here
  url: 'https://osapi-io.github.io/',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/osapi/',
  trailingSlash: false,

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'retr0h', // Usually your GitHub org/user name.
  projectName: 'osapi', // Usually your repo name.

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en']
  },

  presets: [
    [
      'classic',
      {
        docs: {
          routeBasePath: '/',
          sidebarPath: require.resolve('./sidebars.ts'),
          docItemComponent: '@theme/ApiItem' // Derived from docusaurus-theme-openapi
        },
        blog: {},
        pages: {},
        theme: {
          customCss: require.resolve('./src/css/custom.css')
        }
      } satisfies Preset.Options
    ]
  ],

  themeConfig: {
    docs: {
      sidebar: {
        hideable: true
      }
    },
    navbar: {
      logo: {
        alt: 'OSAPI Logo',
        src: 'img/logo.png'
      },
      items: [
        {
          label: 'Getting Started',
          position: 'left',
          to: '/osapi'
        },
        {
          type: 'dropdown',
          label: 'Features',
          position: 'left',
          items: [
            {
              type: 'doc',
              label: 'Node Management',
              docId: 'sidebar/features/node-management'
            },
            {
              type: 'doc',
              label: 'Network Management',
              docId: 'sidebar/features/network-management'
            },
            {
              type: 'doc',
              label: 'System Facts',
              docId: 'sidebar/features/system-facts'
            },
            {
              type: 'doc',
              label: 'Agent Lifecycle',
              docId: 'sidebar/features/agent-lifecycle'
            },
            {
              type: 'doc',
              label: 'Job System',
              docId: 'sidebar/features/job-system'
            },
            {
              type: 'doc',
              label: 'Audit Logging',
              docId: 'sidebar/features/audit-logging'
            },
            {
              type: 'doc',
              label: 'Command Execution',
              docId: 'sidebar/features/command-execution'
            },
            {
              type: 'doc',
              label: 'File Management',
              docId: 'sidebar/features/file-management'
            },
            {
              type: 'doc',
              label: 'Container Management',
              docId: 'sidebar/features/container-management'
            },
            {
              type: 'doc',
              label: 'Cron Management',
              docId: 'sidebar/features/cron-management'
            },
            {
              type: 'doc',
              label: 'Sysctl Management',
              docId: 'sidebar/features/sysctl-management'
            },
            {
              type: 'doc',
              label: 'NTP Management',
              docId: 'sidebar/features/ntp-management'
            },
            {
              type: 'doc',
              label: 'Timezone Management',
              docId: 'sidebar/features/timezone-management'
            },
            {
              type: 'doc',
              label: 'Health Checks',
              docId: 'sidebar/features/health-checks'
            },
            {
              type: 'doc',
              label: 'Authentication & RBAC',
              docId: 'sidebar/features/authentication'
            },
            {
              type: 'doc',
              label: 'Notifications',
              docId: 'sidebar/features/notifications'
            },
            {
              type: 'doc',
              label: 'Distributed Tracing',
              docId: 'sidebar/features/distributed-tracing'
            },
            {
              type: 'doc',
              label: 'Metrics',
              docId: 'sidebar/features/metrics'
            },
            {
              type: 'doc',
              label: 'Power Management',
              docId: 'sidebar/features/power-management'
            }
          ]
        },
        {
          type: 'doc',
          label: 'Usage',
          position: 'left',
          docId: 'sidebar/usage/usage'
        },
        {
          type: 'dropdown',
          label: 'SDK',
          position: 'left',
          items: [
            {
              type: 'html',
              value:
                '<small style="padding: 4px 12px; color: var(--ifm-color-emphasis-600);">Client Library</small>',
              className: 'dropdown-header'
            },
            {
              type: 'doc',
              label: 'Overview',
              docId: 'sidebar/sdk/client/client'
            },
            {
              type: 'html',
              value:
                '<small style="padding: 4px 12px; color: var(--ifm-color-emphasis-600);">Node Info</small>',
              className: 'dropdown-header'
            },
            {
              type: 'doc',
              label: 'Status',
              docId: 'sidebar/sdk/client/status'
            },
            {
              type: 'doc',
              label: 'Hostname',
              docId: 'sidebar/sdk/client/hostname'
            },
            {
              type: 'doc',
              label: 'Disk',
              docId: 'sidebar/sdk/client/disk'
            },
            {
              type: 'doc',
              label: 'Memory',
              docId: 'sidebar/sdk/client/memory'
            },
            {
              type: 'doc',
              label: 'Load',
              docId: 'sidebar/sdk/client/load'
            },
            {
              type: 'doc',
              label: 'Uptime',
              docId: 'sidebar/sdk/client/uptime'
            },
            {
              type: 'doc',
              label: 'OS',
              docId: 'sidebar/sdk/client/os'
            },
            {
              type: 'html',
              value:
                '<small style="padding: 4px 12px; color: var(--ifm-color-emphasis-600);">Network</small>',
              className: 'dropdown-header'
            },
            {
              type: 'doc',
              label: 'DNS',
              docId: 'sidebar/sdk/client/dns'
            },
            {
              type: 'doc',
              label: 'Ping',
              docId: 'sidebar/sdk/client/ping'
            },
            {
              type: 'html',
              value:
                '<small style="padding: 4px 12px; color: var(--ifm-color-emphasis-600);">System Config</small>',
              className: 'dropdown-header'
            },
            {
              type: 'doc',
              label: 'Sysctl',
              docId: 'sidebar/sdk/client/sysctl'
            },
            {
              type: 'doc',
              label: 'NTP',
              docId: 'sidebar/sdk/client/ntp'
            },
            {
              type: 'doc',
              label: 'Timezone',
              docId: 'sidebar/sdk/client/timezone'
            },
            {
              type: 'html',
              value:
                '<small style="padding: 4px 12px; color: var(--ifm-color-emphasis-600);">Operations</small>',
              className: 'dropdown-header'
            },
            {
              type: 'doc',
              label: 'Command',
              docId: 'sidebar/sdk/client/command'
            },
            {
              type: 'doc',
              label: 'Power',
              docId: 'sidebar/sdk/client/power'
            },
            {
              type: 'html',
              value:
                '<small style="padding: 4px 12px; color: var(--ifm-color-emphasis-600);">Containers & Scheduling</small>',
              className: 'dropdown-header'
            },
            {
              type: 'doc',
              label: 'Docker',
              docId: 'sidebar/sdk/client/docker'
            },
            {
              type: 'doc',
              label: 'Cron',
              docId: 'sidebar/sdk/client/cron'
            },
            {
              type: 'html',
              value:
                '<small style="padding: 4px 12px; color: var(--ifm-color-emphasis-600);">Files</small>',
              className: 'dropdown-header'
            },
            {
              type: 'doc',
              label: 'File',
              docId: 'sidebar/sdk/client/file'
            },
            {
              type: 'doc',
              label: 'FileDeploy',
              docId: 'sidebar/sdk/client/file_deploy'
            },
            {
              type: 'html',
              value:
                '<small style="padding: 4px 12px; color: var(--ifm-color-emphasis-600);">Management</small>',
              className: 'dropdown-header'
            },
            {
              type: 'doc',
              label: 'Agent',
              docId: 'sidebar/sdk/client/agent'
            },
            {
              type: 'doc',
              label: 'Job',
              docId: 'sidebar/sdk/client/job'
            },
            {
              type: 'doc',
              label: 'Health',
              docId: 'sidebar/sdk/client/health'
            },
            {
              type: 'doc',
              label: 'Audit',
              docId: 'sidebar/sdk/client/audit'
            },
            {
              type: 'html',
              value: '<hr style="margin: 0.3rem 0;">',
              className: 'dropdown-separator'
            },
            {
              type: 'html',
              value:
                '<small style="padding: 4px 12px; color: var(--ifm-color-emphasis-600);">Platform</small>',
              className: 'dropdown-header'
            },
            {
              type: 'doc',
              label: 'Detection',
              docId: 'sidebar/sdk/platform/detection'
            },
            {
              type: 'html',
              value: '<hr style="margin: 0.3rem 0;">',
              className: 'dropdown-separator'
            },
            {
              type: 'html',
              value:
                '<small style="padding: 4px 12px; color: var(--ifm-color-emphasis-600);">Orchestrator</small>',
              className: 'dropdown-header'
            },
            {
              href: 'https://github.com/osapi-io/osapi-orchestrator',
              label: 'osapi-orchestrator'
            }
          ]
        },
        {
          label: 'API',
          position: 'left',
          to: '/category/api'
        },
        {
          href: 'https://github.com/retr0h/osapi',
          position: 'right',
          className: 'header-github-link',
          'aria-label': 'GitHub repository'
        }
      ]
    },
    footer: undefined,
    colorMode: {
      defaultMode: 'dark',
      disableSwitch: false
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.palenight,
      prism: {
        additionalLanguages: [
          'ruby',
          'csharp',
          'php',
          'java',
          'powershell',
          'json',
          'bash',
          'shell-session'
        ]
      },
      languageTabs: [
        {
          highlight: 'python',
          language: 'python',
          logoClass: 'python'
        },
        {
          highlight: 'bash',
          language: 'curl',
          logoClass: 'bash'
        },
        {
          highlight: 'csharp',
          language: 'csharp',
          logoClass: 'csharp'
        },
        {
          highlight: 'go',
          language: 'go',
          logoClass: 'go'
        },
        {
          highlight: 'javascript',
          language: 'nodejs',
          logoClass: 'nodejs'
        },
        {
          highlight: 'ruby',
          language: 'ruby',
          logoClass: 'ruby'
        },
        {
          highlight: 'php',
          language: 'php',
          logoClass: 'php'
        },
        {
          highlight: 'java',
          language: 'java',
          logoClass: 'java',
          variant: 'unirest'
        },
        {
          highlight: 'powershell',
          language: 'powershell',
          logoClass: 'powershell'
        }
      ]
    }
  } satisfies Preset.ThemeConfig,

  plugins: [
    [
      'docusaurus-plugin-openapi-docs',
      {
        id: 'openapi',
        docsPluginId: 'classic',
        config: {
          osapi: {
            specPath: '../internal/controller/api/gen/api.yaml',
            outputDir: 'docs/gen/api',
            sidebarOptions: {
              groupPathsBy: 'tag',
              categoryLinkSource: 'tag'
            }
          } satisfies OpenApiPlugin.Options
        } satisfies Plugin.PluginOptions
      }
    ]
  ],

  themes: ['docusaurus-theme-openapi-docs', '@docusaurus/theme-mermaid']
};

export default async function createConfig() {
  return config;
}
