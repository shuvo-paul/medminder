# Frontend

## Tech Stack

- Framework: [SvelteKit](https://kit.svelte.dev) (SPA mode, adapter-static)
- Build tool: [Vite](https://vitejs.dev)
- Styling: [Tailwind CSS v4](https://tailwindcss.com)
- Components: [shadcn-svelte](https://www.shadcn-svelte.com)
- Icons: [Lucide Svelte](https://lucide.dev)
- PWA: [vite-plugin-pwa](https://vite-pwa-org.netlify.app)
- Package manager: pnpm
- Build output: `cmd/server/web/dist/` (embedded into Go binary via `go:embed`)

## Commands

| Task | Command | Directory |
|---|---|---|
| Dev server (HMR) | `pnpm dev` | `web/` |
| Production build | `pnpm build` | `web/` |
| Type/syntax check | `pnpm check` | `web/` |
| Preview build | `pnpm preview` | `web/` |
| Install deps | `make web-deps` | root |
| Full prod build | `make embed-frontend` | root |

## Code Style

- Tailwind CSS v4 for all styling — no custom CSS files without good reason
- Use shadcn-svelte components for UI primitives (buttons, dialogs, forms)
- Components follow SvelteKit conventions: `+page.svelte` for routes, `+layout.svelte` for layouts
- SPA mode: client-side routing only, no SSR — `adapter-static` with `fallback: 'index.html'`

## Build & Embed

The frontend builds to `cmd/server/web/dist/` via `adapter-static`. The Go binary embeds this directory using `//go:embed web/dist/*`. Run `make embed-frontend` before production builds to produce a self-contained binary.
