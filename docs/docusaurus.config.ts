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
      title: 'OSAPI',
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
              label: 'System Management',
              docId: 'sidebar/features/system-management'
            },
            {
              type: 'doc',
              label: 'Network Management',
              docId: 'sidebar/features/network-management'
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
              label: 'Distributed Tracing',
              docId: 'sidebar/features/distributed-tracing'
            },
            {
              type: 'doc',
              label: 'Metrics',
              docId: 'sidebar/features/metrics'
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
          label: 'API',
          position: 'left',
          to: '/category/api'
        },
        {
          type: 'docsVersionDropdown',
          position: 'right'
        },
        {
          href: 'https://github.com/retr0h/osapi',
          position: 'right',
          className: 'header-github-link',
          'aria-label': 'GitHub repository'
        }
      ]
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Community',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/retr0h/osapi'
            }
          ]
        }
      ],
      copyright: `Copyright © ${new Date().getFullYear()} <a href="https://github.com/retr0h">@retr0h</a>`
    },
    colorMode: {
      defaultMode: 'dark',
      disableSwitch: false
    },
    prism: {
      theme: prismThemes.palenight,
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
          system: {
            specPath: '../internal/client/gen/api.yaml',
            outputDir: 'docs/gen/api',
            downloadUrl:
              'https://github.com/retr0h/osapi/blob/main/internal/client/gen/api.yaml',
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
