<script lang="ts">
	import { page } from '$app/stores';

	let isResending = $state(false);
	let visible = $state(false);

	const DISMISS_KEY = 'email_verification_banner_dismissed';
	const DISMISS_EXPIRY_MS = 24 * 60 * 60 * 1000; // 24 hours

	function isDismissed(): boolean {
		if (typeof localStorage === 'undefined') return false;
		const raw = localStorage.getItem(DISMISS_KEY);
		if (!raw) return false;
		const expiry = parseInt(raw, 10);
		if (isNaN(expiry)) return false;
		return Date.now() < expiry;
	}

	function dismiss() {
		if (typeof localStorage !== 'undefined') {
			localStorage.setItem(DISMISS_KEY, String(Date.now() + DISMISS_EXPIRY_MS));
		}
		visible = false;
	}

	$effect(() => {
		if (typeof window === 'undefined') return;

		const path = $page.url.pathname;

		// Track localStorage reactively via storage event
		const token = localStorage.getItem('access_token');
		const emailVerified = localStorage.getItem('email_verified');

		// Never show on login/register pages
		if (path === '/login' || path === '/register') {
			visible = false;
			return;
		}

		// Show banner only when: has token, email not verified, not dismissed
		if (token && emailVerified === 'false' && !isDismissed()) {
			visible = true;
		} else {
			visible = false;
		}

		// Re-run when localStorage changes (e.g., after login)
		function handleStorageChange() {
			const newToken = localStorage.getItem('access_token');
			const newEmailVerified = localStorage.getItem('email_verified');
			if (newToken && newEmailVerified === 'false' && !isDismissed()) {
				visible = true;
			} else {
				visible = false;
			}
		}

		window.addEventListener('storage', handleStorageChange);
		return () => window.removeEventListener('storage', handleStorageChange);
	});

	async function resendVerification(e: Event) {
		e.preventDefault();
		const token = localStorage.getItem('access_token');
		if (!token) return;

		isResending = true;
		try {
			const response = await fetch('/api/auth/email/resend-verification', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${token}`,
				},
				body: JSON.stringify({}),
			});

			if (response.ok) {
				dismiss();
			}
		} finally {
			isResending = false;
		}
	}
</script>

{#if visible}
	<div class="fixed left-0 right-0 top-12 z-40 px-4 py-2">
		<div class="mx-auto flex max-w-md items-center gap-3 rounded-lg bg-destructive/10 border border-destructive/20 px-3 py-2 text-xs">
			<span class="flex-1 truncate font-medium text-destructive">Email not verified</span>
			<span class="shrink-0 text-muted-foreground">Check your inbox or</span>
			<button
				onclick={resendVerification}
				disabled={isResending}
				class="shrink-0 text-sm font-medium text-primary hover:underline disabled:opacity-50"
			>
				{isResending ? 'Sending...' : 'Resend'}
			</button>
			<button
				onclick={dismiss}
				class="shrink-0 text-muted-foreground hover:text-foreground"
				aria-label="Dismiss for 24h"
			>
				<svg class="h-3.5 w-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
					<path d="M18 6 6 18M6 6l12 12" />
				</svg>
			</button>
		</div>
	</div>
{/if}