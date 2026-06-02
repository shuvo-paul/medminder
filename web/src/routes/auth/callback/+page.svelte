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

		console.log('[OAuth callback] page loaded', { code: !!code, state: !!state, stateLen: state?.length, search: window.location.search });

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
			console.log('[OAuth callback] POST /api/auth/oauth/token', { codeLen: code.length, stateLen: state.length });
			const res = await fetch('/api/auth/oauth/token', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body,
			});

			console.log('[OAuth callback] POST response', { status: res.status, ok: res.ok });

			if (!res.ok) {
				const data = await res.json();
				console.log('[OAuth callback] POST error body', data);
				let redirect = '/login';
				try {
					const decoded = base64UrlToStandard(state);
					console.log('[OAuth callback] decoding state', { decoded: decoded.slice(0, 40) + '...' });
					const parsed = JSON.parse(atob(decoded));
					console.log('[OAuth callback] state parsed', parsed);
					redirect = parsed.redirect || '/login';
				} catch (e) {
					console.log('[OAuth callback] state parse failed', e);
				}
				const errCode = data.title || data.detail || 'unknown_error';
				console.log('[OAuth callback] redirecting to error', `${redirect}?oauth_error=${errCode}&provider=google`);
				goto(`${redirect}?oauth_error=${errCode}&provider=google`);
				return;
			}

			const data = await res.json();
			console.log('[OAuth callback] token exchange success', { hasAccessToken: !!data.access_token, user: data.user });

			localStorage.setItem('access_token', data.access_token);
			localStorage.setItem('refresh_token', data.refresh_token);
			localStorage.setItem('email_verified', String(data.user.email_verified));

			console.log('[OAuth callback] tokens stored, verifying...', { storedToken: localStorage.getItem('access_token')?.slice(0, 20) + '...' });

			let redirect = '/dashboard';
			try {
				const decoded = base64UrlToStandard(state);
				const parsed = JSON.parse(atob(decoded));
				console.log('[OAuth callback] state parsed for redirect', parsed);
				redirect = parsed.redirect || '/dashboard';
			} catch (e) {
				console.log('[OAuth callback] state parse failed on success path', e);
			}
			console.log('[OAuth callback] navigating to', redirect);
			goto(redirect);
		} catch (e) {
			console.log('[OAuth callback] network error', e);
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
