// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

import type * as Preset from '@docusaurus/preset-classic';
import type { Config } from '@docusaurus/types';
import type * as Plugin from '@docusaurus/types/src/plugin';
import type * as OpenApiPlugin from 'docusaurus-plugin-openapi-docs';
import type { PrismTheme } from 'prism-react-renderer';
import { palette } from './src/palette.gen';
import { behaviorItems, featureItems, sdkItems } from './src/navbar';

// codeTheme is a Prism theme from the CLI's palette, so a code block reads as
// the terminal output it shows: keywords magenta, strings yellow, functions
// and properties cyan, numbers the logo blue, comments gray.
function codeTheme(
  p: (typeof palette)['dark'] | (typeof palette)['light'],
  text: string,
  background: string
): PrismTheme {
  const [cyan, magenta, yellow, , blue] = p.headings;
  return {
    plain: { color: text, backgroundColor: background },
    styles: [
      {
        types: ['comment', 'prolog', 'doctype', 'cdata'],
        style: { color: p.gray, fontStyle: 'italic' }
      },
      {
        types: ['keyword', 'builtin', 'important', 'atrule'],
        style: { color: magenta }
      },
      {
        types: ['string', 'char', 'attr-value', 'regex', 'inserted'],
        style: { color: yellow }
      },
      {
        types: [
          'function',
          'class-name',
          'property',
          'attr-name',
          'tag',
          'selector'
        ],
        style: { color: cyan }
      },
      {
        types: ['number', 'boolean', 'constant', 'symbol', 'variable'],
        style: { color: blue }
      },
      { types: ['punctuation', 'operator'], style: { color: p.gray } },
      // bash: a flag, such as -r, and an argument of an assignment.
      { types: ['parameter', 'assign-left'], style: { color: blue } },
      { types: ['deleted'], style: { color: '#ef4444' } }
    ]
  };
}

const config: Config = {
  title: 'OSAPI',
  tagline: 'Typed, repeatable operations for Linux hosts.',
  favicon: 'img/favicon.ico',
  markdown: {
    mermaid: true,
    format: 'detect',
    hooks: {
      onBrokenMarkdownLinks: 'warn'
    }
  },

  // Set the production url of your site here
  url: 'https://osapi-io.github.io/',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/osapi/',
  trailingSlash: false,

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'osapi-io', // Usually your GitHub org/user name.
  projectName: 'osapi', // Usually your repo name.

  onBrokenLinks: 'throw',

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
          customCss: [
            require.resolve('./src/css/palette.gen.css'),
            require.resolve('./src/css/custom.css')
          ]
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
        // The block logo of the README, asset/logo-*.svg: light on a light
        // page, dark on a dark one.
        src: 'img/logo-light.svg',
        srcDark: 'img/logo-dark.svg'
      },
      items: [
        {
          label: 'Getting Started',
          position: 'left',
          to: '/osapi',
          // Active on the home page only: '/osapi' starts every URL.
          activeBaseRegex: '^/osapi/?$'
        },
        {
          type: 'dropdown',
          label: 'Features',
          position: 'left',
          items: featureItems()
        },
        {
          // A link, not a doc item: a doc item is active on each page of
          // its sidebar, which is every page.
          label: 'Usage',
          position: 'left',
          to: '/osapi/sidebar/usage/',
          activeBaseRegex: '/sidebar/usage/'
        },
        {
          type: 'dropdown',
          label: 'SDK',
          position: 'left',
          items: sdkItems()
        },
        {
          type: 'dropdown',
          label: 'Behaviors',
          position: 'left',
          items: behaviorItems()
        },
        {
          label: 'API',
          position: 'left',
          to: '/category/api'
        },
        {
          href: 'https://github.com/osapi-io/osapi',
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
      theme: codeTheme(palette.light, '#18202b', '#f8f9fb'),
      darkTheme: codeTheme(palette.dark, '#e2e8f0', '#000000'),
      // At this level, not in a prism key of its own, which Docusaurus
      // does not read: bash and the rest were never highlighted.
      additionalLanguages: [
        'ruby',
        'csharp',
        'php',
        'java',
        'powershell',
        'json',
        'bash',
        'shell-session'
      ],
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
      '@docusaurus/plugin-client-redirects',
      {
        // api-guidelines.md and principles.md were contributor documents in an
        // operator's navigation. Their rules are now stated once in the
        // specifications repository (005-building-a-domain, FR-014, FR-015,
        // FR-022), and these addresses keep working for anyone holding a link.
        // adding-an-api-domain became the behavior tutorial.
        redirects: [
          {
            from: '/sidebar/architecture/api-guidelines',
            to: '/sidebar/development/your-first-behavior'
          },
          {
            from: '/sidebar/architecture/principles',
            to: '/sidebar/development/your-first-behavior'
          },
          {
            from: '/sidebar/development/adding-an-api-domain',
            to: '/sidebar/development/your-first-behavior'
          },
          {
            from: '/sidebar/architecture/job-architecture',
            to: '/sidebar/architecture/request-flow'
          },
          {
            from: '/sidebar/architecture/system-architecture',
            to: '/sidebar/architecture/components'
          }
        ]
      }
    ],
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
