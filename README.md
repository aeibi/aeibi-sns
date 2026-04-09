<div align="center">
  <img src="./docs/images/aeibi.svg" alt="AeiBi Logo" width="180" />
  <h1>AeiBi</h1>
  <p>
    A lightweight community platform focused on posting, discussion,
    relationships, and inbox notifications.
  </p>
  <p>
    <a href="https://aeibi.com">
      <img alt="Live Demo" src="https://img.shields.io/badge/Live%20Demo-aeibi.com-0ea5e9?logo=vercel&logoColor=white&style=flat-square" />
    </a>
    <a href="https://github.com/aeibi/aeibi-sns/graphs/contributors">
      <img alt="Contributors" src="https://img.shields.io/github/contributors/aeibi/aeibi-sns.svg?style=flat-square" />
    </a>
    <a href="https://github.com/aeibi/aeibi-sns/network/members">
      <img alt="Forks" src="https://img.shields.io/github/forks/aeibi/aeibi-sns.svg?style=flat-square" />
    </a>
    <a href="https://github.com/aeibi/aeibi-sns/stargazers">
      <img alt="Stars" src="https://img.shields.io/github/stars/aeibi/aeibi-sns.svg?style=flat-square" />
    </a>
    <img alt="Status" src="https://img.shields.io/badge/Status-Early%20Stage-f59e0b?style=flat-square" />
    <img alt="Backend" src="https://img.shields.io/badge/Backend-Go%201.25-00ADD8?logo=go&logoColor=white&style=flat-square" />
    <img alt="Frontend" src="https://img.shields.io/badge/Frontend-React%2019-61DAFB?logo=react&logoColor=0A0A0A&style=flat-square" />
    <img alt="Database" src="https://img.shields.io/badge/Database-PostgreSQL-4169E1?logo=postgresql&logoColor=white&style=flat-square" />
    <img alt="API" src="https://img.shields.io/badge/API-gRPC-244c5a?logo=grpc&logoColor=white&style=flat-square" />
  </p>
</div>

![AeiBi Home Screenshot](./docs/images/home.png)
> Live Demo: https://aeibi.com

## Project Status

The project is in an early stage. Core community flows are already in place, and features are being actively iterated.

## Features

- Account system: sign up, log in, token refresh, logout, profile updates, password change
- Content publishing: create posts (text, images, tags), edit/delete posts, public/private visibility
- Social interactions: likes, collections, comments, replies, comment likes
- Relationship graph: follow/unfollow, followers/following lists, relation search
- Inbox center: follow and comment notifications, unread counts, mark all as read, archive single messages
- Search & discovery: post search, tag search, user search, tag/user prefix suggestions
- Moderation: report posts, comments, and users
- File service: upload files, query metadata, retrieve file content (S3-compatible object storage)

## Local Development

Prerequisites:
- Go `1.25.4+`
- Docker Engine `28+`
- Docker Compose `v2+`
- Node.js `22.x` (LTS recommended)
- pnpm `10+`

### Startup Modes

Start required dependencies first (from repository root):

```bash
docker compose -f docker/docker-compose.yaml up -d
```

Mode 1: Frontend dev server + backend-only API

Frontend:

```bash
cd web
pnpm install
pnpm run dev
```

Backend:

```bash
go run ./cmd backend --config ./config.example.yaml
```

Mode 2: Embedded frontend release + full backend

Build frontend release assets first (from `web`):

```bash
cd web
pnpm install
pnpm run release
cd ..
```

Start full service:

```bash
go run ./cmd --config ./config.example.yaml
```

Notes:

- In Mode 1, frontend is served by Vite dev server; backend serves API routes only (`/api/*` and `/file/*`).
- In Mode 2, backend serves embedded frontend assets from `web/dist`.

## Star History

<a href="https://www.star-history.com/?repos=aeibi%2Faeibi-sns&type=date&legend=bottom-right">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=aeibi/aeibi-sns&type=date&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=aeibi/aeibi-sns&type=date&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=aeibi/aeibi-sns&type=date&legend=top-left" />
 </picture>
</a>

## License

MIT

## Contributing

Contributions of all kinds are welcome, such as bug fixes, new features, documentation improvements, etc.
