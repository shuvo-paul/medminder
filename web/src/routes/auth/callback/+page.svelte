<script lang="ts">
	import { goto } from '$app/navigation';
	import { LoaderCircle } from '@lucide/svelte';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';

	let error = $state('');
	let isLoading = $state(true);

	function base64UrlToStandard(str: string): string {
		str = str.replace(/-/g, '+').replace(/_/g, '/');
		const padding = str.length % 4;
		if (padding) str += '='.repeat(4 - padding);
		return str;
	}

	if (typeof window !== 'undefined') {
		const params = new URLSearchParams(window.location.search);
		const code = params.get('code');
		const state = params.get('state');

		if (!code || !state) {
			error = 'Invalid callback parameters.';
			isLoading = false;
		} else {
			exchangeToken(code, state);
		}
	}

	async function exchangeToken(code: string, state: string) {
		try {
			const body = JSON.stringify({ code, state });
			const res = await fetch('/api/auth/oauth/token', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body,
			});

			if (!res.ok) {
				const data = await res.json();
				let redirect = '/login';
				let email = '';
				try {
					const decoded = base64UrlToStandard(state);
					const parsed = JSON.parse(atob(decoded));
					redirect = parsed.redirect || '/login';
				} catch (e) {
					// State unreadable — fall back to /login.
				}
				const errCode = data.detail || data.title || 'unknown_error';

				// Try to extract email from errors for email_exists
				if (errCode === 'email_exists') {
					try {
						const msg = data.errors?.[0]?.message;
						if (msg) {
							const parsed = JSON.parse(msg);
							email = parsed.email || '';
						}
					} catch {
						// Could not extract email from error
					}
				}

				const params = new URLSearchParams({ oauth_error: errCode, provider: 'google' });
				if (email) params.set('email', email);

				if (errCode === 'link_failed') {
					const errorMsg = data.errors?.[0]?.message;
					if (errorMsg) params.set('oauth_message', errorMsg);
				}
				goto(`${redirect}?${params.toString()}`);
				return;
			}

			const data = await res.json();

			localStorage.setItem('access_token', data.access_token);
			localStorage.setItem('refresh_token', data.refresh_token);
			localStorage.setItem('email_verified', String(data.user.email_verified));

			let redirect = '/dashboard';
			try {
				const decoded = base64UrlToStandard(state);
				const parsed = JSON.parse(atob(decoded));
				redirect = parsed.redirect || '/dashboard';
			} catch (e) {
				// State unreadable — fall back to /dashboard.
			}
			goto(redirect);
		} catch (e) {
			goto('/login?oauth_error=network_error');
		}
	}
</script>

<div class="flex min-h-screen flex-col items-center justify-center px-4">
	<Card class="w-full max-w-sm">
		<CardHeader>
			<CardTitle class="text-center">
				{#if error}
					Callback Error
				{:else}
					Connecting
				{/if}
			</CardTitle>
			<CardDescription class="text-center">
				{#if error}
					{error}
				{:else}
					Please wait while we sign you in
				{/if}
			</CardDescription>
		</CardHeader>
		<CardContent class="flex justify-center pb-6">
			{#if !error}
				<LoaderCircle class="h-8 w-8 animate-spin text-muted-foreground" />
			{/if}
		</CardContent>
	</Card>
</div>
