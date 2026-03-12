# Journal Frontend

Next.js App Router frontend for the main-site shell, paper archive, auth flows, and search controls.

## Runtime Notes

- Default local URL: `http://127.0.0.1:3000`
- If `JOURNAL_API_ORIGIN` is unset, server-side fetches fall back to same-origin `/api/v1`
- When the backend stack is unavailable, `/papers` intentionally renders a diagnosable error state instead of crashing

## Local Development

```bash
npm install
npm run dev -- --hostname 127.0.0.1 --port 3000
```

Useful pages:

- `/`
- `/papers`
- `/papers?query=人工智能论文&sort=relevance&engine=fulltext&shadow_compare=true`
- `/login`
- `/register`

## Commands

```bash
npm run lint
npx next build --webpack
npm run smoke
```

What each command does:

- `npm run lint`: static linting for the App Router codebase
- `npx next build --webpack`: production build and TypeScript validation
- `npm run smoke`: starts `next start` on `127.0.0.1:3100` and asserts `/`, `/papers`, and query-mode `/papers` render the expected shell and fallback UI

## Search Verification

The `/papers` page is URL-driven. These query params are part of the supported surface:

- `query`
- `discipline`
- `sort=relevance|newest|quality`
- `engine=auto|hybrid|fulltext`
- `shadow_compare=true`

If you need live search results instead of fallback UI, start the backend `journal-api` and `paper-rpc` stack before opening the page.

## CI Hooks

The CI workflow uses the same commands as local validation:

- backend: `go test ./api/... ./rpc/paper/... ./model`
- backend contract: `make api-contract-test`
- backend search benchmark: `make search-bench`
- frontend: `npm run lint`
- frontend build: `npx next build --webpack`
- frontend smoke: `npm run smoke`
