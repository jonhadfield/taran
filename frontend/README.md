# MailBrief Frontend

Next.js 16 dashboard application providing authentication, email browsing, digest viewing, analytics, and account management.

## Tech Stack

- **Next.js 16** with App Router and React 19
- **Better Auth** for OAuth (Google + GitHub)
- **Tailwind CSS 4** + **shadcn/ui** + **Radix UI** for styling
- **Prisma** as the Better Auth database adapter
- **Vitest** + **Testing Library** for unit tests
- **Playwright** for E2E tests

## Architecture

```
src/
├── app/                    App Router routes
│   ├── (auth)/             Auth pages (login, not-invited)
│   ├── (dashboard)/        Protected pages (all main views)
│   ├── shared/[token]/     Public digest sharing
│   └── api/
│       ├── auth/[...all]/  Better Auth handler
│       └── proxy/[...path]/ Reverse proxy to Go backend
│
├── components/
│   ├── ui/                 shadcn/ui primitives
│   └── *.tsx               Feature components
│
├── hooks/                  Custom React hooks
├── lib/                    Utilities, API client, auth config
├── types/                  TypeScript API contracts
├── test/                   Test setup
└── proxy.ts                Next.js 16 request proxy (replaces middleware.ts)
```

## Pages & Routes

### Auth (unauthenticated)
| Route | Page | Description |
|-------|------|-------------|
| `/login` | Login | OAuth sign-in (Google/GitHub) |
| `/not-invited` | Access Denied | Shown when user lacks an invite |

### Dashboard (authenticated)
| Route | Page | Description |
|-------|------|-------------|
| `/` | Dashboard | Overview with analytics, recent emails, latest digest |
| `/onboarding` | Onboarding | 3-step wizard (account setup, forwarding guide, first email) |
| `/inbox` | Inbox | Email list with search, filters, bulk actions, keyboard shortcuts |
| `/inbox/[id]` | Email Detail | Full email with extraction, thread view, actions |
| `/digests` | Digests | Digest list with on-demand generation |
| `/digests/[id]` | Digest Detail | Full digest with comparison and sharing |
| `/senders` | Senders | Sender list with categorisation and analytics |
| `/senders/[address]` | Sender Detail | Per-sender stats (frequency, last seen, emails) |
| `/senders/subscriptions` | Subscriptions | Unsubscribe link tracking and management |
| `/settings` | Settings | Account, digest prefs, filters, API keys, theme, quiet hours |
| `/admin` | Admin | User management, failed emails, waitlist, pipeline health |

### Public
| Route | Page | Description |
|-------|------|-------------|
| `/shared/[token]` | Shared Digest | Public read-only digest view |

## Auth Flow

Better Auth runs entirely in the Next.js layer and manages the `user`, `session`, and `account` tables in the shared PostgreSQL database.

```
Browser → proxy.ts (session check, redirects)
       → /api/auth/* (Better Auth handles OAuth callbacks, session management)
       → /api/proxy/* (proxies to Go backend with session token)
```

**proxy.ts** intercepts every request (except static assets and `/api` routes):
- Authenticated users are redirected away from `/login`
- Unauthenticated users are redirected to `/login` (except `/shared/*`)

**Session config:** 30-day expiry, 24-hour refresh interval, 1-hour freshness window.

## API Client

Two API layers, both routing through the Go backend:

### Client-side (`src/lib/api.ts`)

Used in client components. Requests go through the Next.js proxy route:

```
Browser → /api/proxy/{path} → Go backend /api/{path}
```

Exports: `apiGet`, `apiPost`, `apiPut`, `apiPatch`, `apiDelete`, `apiDeleteJSON`

The proxy route extracts the Better Auth session token from the cookie, strips the HMAC signature, and forwards it as a Bearer token.

### Server-side (`src/lib/server-api.ts`)

Used in server components and server actions. Calls the Go backend directly:

```
Server Component → BACKEND_URL/api/{path}
```

Exports: `serverFetch`

Reads the session token from cookies, strips HMAC, sends as Bearer token. All requests use `cache: "no-store"`.

## UI System

### Styling

- **Tailwind CSS 4** via `@tailwindcss/postcss`
- **shadcn/ui** components: button, card, input, badge, dialog, dropdown-menu, sheet, tabs, switch, scroll-area, skeleton, separator, avatar, and more
- **cn()** utility (clsx + tailwind-merge) for conditional classes

### Theming

Two layers of theming:

1. **Dark/Light mode** via `next-themes` with system detection
2. **Colour themes** via custom `ColorThemeProvider` -- neutral, blue, rose, green, violet, amber
   - Applied as `data-theme` attribute on `<html>`
   - Persisted in cookie + synced to user preferences API

### Layout

- **Desktop (lg+):** Fixed sidebar navigation + header + content area
- **Mobile:** Collapsible sidebar via Sheet component, hamburger menu in header
- Sidebar links: Inbox, Digests, Senders, Settings, Admin (if admin)

### Notifications

- **Toast:** `sonner` for in-app notifications
- **Browser:** Notification API for new email alerts (permission prompted on first inbox visit)

## Key Hooks

| Hook | Purpose |
|------|---------|
| `authClient.useSession()` | Current session and user data |
| `useTheme()` | Dark/light mode toggle (next-themes) |
| `useColorTheme()` | Custom colour theme selection |
| `usePolling(path, initialData, interval)` | Poll an API endpoint at intervals |
| `useEmailNotifications()` | Browser notification permission and dispatch |

## Testing

### Unit Tests (Vitest)

```bash
npm run test         # Watch mode
npm run test:run     # Single run (CI)
```

Configuration: jsdom environment, `@testing-library/jest-dom` matchers, `@vitejs/plugin-react` for JSX.

Test files live alongside their source in `__tests__/` directories:
- `src/lib/__tests__/` -- API client tests
- `src/hooks/__tests__/` -- Hook tests
- `src/components/__tests__/` -- Component tests
- `src/app/(dashboard)/inbox/__tests__/` -- Page-level tests

### E2E Tests (Playwright)

```bash
npm run test:e2e     # Run headless
npm run test:e2e:ui  # Interactive UI mode
```

Configuration: Chromium only, auto-starts dev server on `:3002`, 30s timeout, screenshots on failure.

Requires `BETTER_AUTH_SECRET` in `.env.local`: the tests forge session cookies and must sign them with the same secret the app verifies with. They fail fast if it is unset.

Test files in `e2e/` directory.

## Development

```bash
npm run dev          # Start dev server on :3002
npm run build        # Production build
npm run lint         # ESLint
```

Or via the root Makefile:

```bash
make start-frontend  # Start dev server in background
make stop-frontend   # Stop
```

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `NEXT_PUBLIC_APP_URL` | Public app URL (default: `http://localhost:3002`) |
| `DATABASE_URL` | PostgreSQL connection string (for Prisma / Better Auth) |
| `BETTER_AUTH_SECRET` | Better Auth encryption secret |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | Google OAuth credentials |
| `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | GitHub OAuth credentials |
| `BACKEND_URL` | Go backend URL (default: `http://localhost:8080`) |
| `API_KEY` | API key sent to Go backend |
| `ADMIN_EMAILS` | Comma-separated admin email addresses |
