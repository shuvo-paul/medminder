# Login Page Design — MedMinder

**Date**: 2026-04-21  
**Status**: Approved

---

## Layout & Structure

**Split screen** — left panel ~45%, right panel ~55%.

- **Left panel**: Calming health/wellness photo with logo and tagline ("Never miss a dose") overlaid at the bottom. Photo uses a darkened overlay so text is legible.
- **Right panel**: White/light gray form area. Email field, password field, "Remember me" checkbox (left), "Forgot password?" link (right), "Continue" button, divider with Google/Apple social options.
- **Mobile** (<768px): Stack vertically — illustration on top, form scrolls below. Bottom nav remains visible.

---

## Visual Identity

- **Primary color**: Teal `#0d9488` — fresh, health-forward, trustworthy
- **Left panel**: Darkened health/wellness photo with teal-tinted overlay; logo + tagline at bottom
- **Right panel background**: `#f8f7f4` warm light gray
- **Typography**: System font stack (matches existing app)
- **Spacing**: Generous, approachable

---

## Form Fields & Behavior

- **Fields**: Email, Password
- **Remember me**: Checkbox, stored in localStorage. Pre-fills on return if token valid.
- **Forgot password**: Link — shows toast "coming soon" (not yet implemented)
- **Submit**: Loading spinner on button, errors inline below form
- **Success**: Store `access_token` and `refresh_token` in localStorage, redirect to `/`
- **New user CTA**: "Don't have an account? Sign up" → links to `/register`
- **Social login**: Google/Apple buttons as placeholders (not yet functional)

---

## Components

- `web/src/routes/login/+page.svelte` — main login page
- Reuses existing `web/src/lib/components/ui/` components (button, input, label)
- Uses `lucide-svelte` icons (Eye/EyeOff for password toggle)
- No new shared components needed

---

## Technical Approach

- SvelteKit page with form actions disabled (pure client-side POST to `/api/auth/login`)
- On load: check localStorage for existing valid token → if present, redirect to `/`
- On submit: `POST /api/auth/login` with `{ email, password }`, handle 200 (store tokens + redirect) or 4xx/5xx (show error)
- Password visibility toggle using browser `<input type="password/text">` swap
- No server-side rendering for this page (`+page.ts` with `ssr: false`)

---

## Out of Scope

- Password reset flow (placeholder only)
- Social login integration (buttons present but non-functional)
- Remember me server-side token refresh