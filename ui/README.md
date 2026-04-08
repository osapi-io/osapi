[![license](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](LICENSE)
[![conventional commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)
[![React](https://img.shields.io/badge/React_19-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)](https://react.dev)
[![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?style=for-the-badge&logo=typescript&logoColor=white)](https://www.typescriptlang.org)
[![Vite](https://img.shields.io/badge/Vite-646CFF?style=for-the-badge&logo=vite&logoColor=white)](https://vite.dev)
[![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-06B6D4?style=for-the-badge&logo=tailwindcss&logoColor=white)](https://tailwindcss.com)
[![built with just](https://img.shields.io/badge/Built_with-Just-black?style=for-the-badge&logo=just&logoColor=white)](https://just.systems)
![gitHub commit activity](https://img.shields.io/github/commit-activity/m/osapi-io/osapi-ui?style=for-the-badge)

# OSAPI UI

A React management dashboard for [OSAPI][] with a meshtastic-inspired
design language.

## Screenshots

<p align="center">
  <a href="asset/dashboard.png"><img src="asset/dashboard.png" width="400" alt="Dashboard"></a>
  <a href="asset/configure.png"><img src="asset/configure.png" width="400" alt="Configure"></a>
</p>

## ✨ Features

| Feature | Description |
| --- | --- |
| Dashboard | Fleet health with controller/NATS components, streams, KV stores, object store, and agent cards |
| Configure | Block-based operations builder with per-block target selection and result rendering |
| Auth & RBAC | JWT sign-in with role-based permission gating (Admin, Operator, Viewer) |
| Agent Management | Drain/undrain agents from the dashboard with RBAC-gated controls |
| @fact. References | Auto-complete fact references in DNS and network fields from live API |
| Generated SDK | Typed fetch functions from OSAPI's OpenAPI spec via [orval](https://orval.dev/) |

[OSAPI]: https://github.com/osapi-io/osapi

## 🤝 Contributing

See the [Development](docs/development.md) guide for prerequisites, setup,
and conventions. See the [Contributing](docs/contributing.md) guide before
submitting a PR.

## 📄 License

The [MIT][] License.

[MIT]: LICENSE
