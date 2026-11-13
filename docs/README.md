# Documentation site

The osapi documentation site, built with [Docusaurus](https://docusaurus.io/)
and published to GitHub Pages. Run everything from the repository root through
the shared `docs` recipes, not with `yarn` or `npm` in this directory:

```bash
just docs-start      # Development server on http://localhost:3001
just docs-build      # Production build into docs/build
just docs-generate   # Regenerate the API reference from the OpenAPI spec
just docs-fmt        # Format the site with Prettier
```

`just --list` shows the rest under `[docs]`. See the
[Documentation section](../CONTRIBUTING.md#documentation) of the contributing
guide for how pages are organized.
