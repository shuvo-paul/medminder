<script lang="ts">
	import { page } from '$app/stores';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let status = $state<'loading' | 'success' | 'error'>('loading');
	let errorMessage = $state('');
	let isResending = $state(false);
	let resendSuccess = $state(false);

	onMount(() => {
		const token = $page.url.searchParams.get('token');

		if (!token) {
			status = 'error';
			errorMessage = 'Verification link is invalid. Please check your email and click the link again.';
			return;
		}

		verifyEmail(token);
	});

	async function verifyEmail(token: string) {
		try {
			const response = await fetch('/api/auth/email/verify', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ token }),
			});

			if (!response.ok) {
				let detail = 'Verification failed. The link may have expired.';
				try {
					const data = await response.json();
					detail = data.detail || detail;
				} catch {
					// Non-JSON response — use default message
				}

				if (response.status === 404 || detail.includes('not found') || detail.includes('invalid')) {
					errorMessage = 'This verification link is invalid or has expired. Please request a new one.';
				} else if (response.status === 400) {
					errorMessage = detail;
				} else if (response.status === 409) {
					errorMessage = detail;
				} else {
					errorMessage = 'Something went wrong. Please try again later.';
				}
				status = 'error';
				return;
			}

			let data;
			try {
				data = await response.json();
			} catch {
				errorMessage = 'Invalid response from server. Please try again.';
				status = 'error';
				return;
			}
			localStorage.setItem('access_token', data.access_token);
			localStorage.setItem('email_verified', 'true');
			localStorage.removeItem('refresh_token');
			status = 'success';

			setTimeout(() => {
				goto('/');
			}, 2000);
		} catch {
			errorMessage = 'Network error. Please check your connection and try again.';
			status = 'error';
		}
	}

	async function resendVerification() {
		const token = localStorage.getItem('access_token');
		if (!token) {
			goto('/login');
			return;
		}

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
				resendSuccess = true;
			} else {
				const data = await response.json();
				errorMessage = data.detail || 'Failed to resend verification email.';
				status = 'error';
			}
		} catch {
			errorMessage = 'Network error. Please check your connection.';
			status = 'error';
		} finally {
			isResending = false;
		}
	}
</script>

<div class="flex min-h-screen flex-col items-center justify-center bg-background px-4">
	<div class="mx-auto w-full max-w-sm">
		{#if status === 'loading'}
			<Card>
				<CardHeader>
					<CardTitle class="text-center">Verifying your email...</CardTitle>
					<CardDescription class="text-center">Please wait while we verify your email address.</CardDescription>
				</CardHeader>
				<CardContent class="flex justify-center py-6">
					<div class="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
				</CardContent>
			</Card>
		{:else if status === 'success'}
			<Card>
				<CardHeader>
					<CardTitle class="text-center text-green-600 dark:text-green-400">Email Verified!</CardTitle>
					<CardDescription class="text-center">Your email has been successfully verified.</CardDescription>
				</CardHeader>
				<CardContent class="space-y-4 text-center">
					<p class="text-sm text-muted-foreground">Redirecting you to the home page...</p>
					<div class="flex justify-center">
						<div class="h-8 w-8 animate-spin rounded-full border-4 border-green-500 border-t-transparent"></div>
					</div>
				</CardContent>
			</Card>
		{:else if status === 'error'}
			<Card>
				<CardHeader>
					<CardTitle class="text-center text-destructive">Verification Failed</CardTitle>
					<CardDescription class="text-center">We couldn't verify your email address.</CardDescription>
				</CardHeader>
				<CardContent class="space-y-4">
					<div class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
						{errorMessage}
					</div>

					{#if resendSuccess}
						<div class="rounded-md bg-green-500/10 p-3 text-sm text-green-600 dark:text-green-400">
							Verification email sent! Check your inbox.
						</div>
					{:else}
						<Button variant="outline" class="w-full" onclick={resendVerification} disabled={isResending}>
							{#if isResending}
								Resending...
							{:else}
								Resend verification email
							{/if}
						</Button>
					{/if}

					<p class="text-center text-sm text-muted-foreground">
						<a href="/login" class="text-primary hover:underline">Back to sign in</a>
					</p>
				</CardContent>
			</Card>
		{/if}
	</div>
</div>
