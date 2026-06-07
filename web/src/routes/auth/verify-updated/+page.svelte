<script lang="ts">
	import { page } from '$app/stores';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let status = $state<'loading' | 'success' | 'error'>('loading');
	let errorMessage = $state('');

	onMount(() => {
		const token = $page.url.searchParams.get('token');

		if (!token) {
			status = 'error';
			errorMessage = 'Verification link is invalid. Please check your email and click the link again.';
			return;
		}

		verifyUpdatedEmail(token);
	});

	async function verifyUpdatedEmail(token: string) {
		try {
			const response = await fetch('/api/auth/email/verify-updated', {
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

				if (response.status === 404 || detail.includes('not found') || detail.includes('invalid') || detail.includes('expired')) {
					errorMessage = 'This verification link is invalid or has expired.';
				} else if (response.status === 400) {
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
				goto('/account');
			}, 2000);
		} catch {
			errorMessage = 'Network error. Please check your connection and try again.';
			status = 'error';
		}
	}
</script>

<div class="flex min-h-screen flex-col items-center justify-center bg-background px-4">
	<div class="mx-auto w-full max-w-sm">
		{#if status === 'loading'}
			<Card>
				<CardHeader>
					<CardTitle class="text-center">Updating your email...</CardTitle>
					<CardDescription class="text-center">Please wait while we update your email address.</CardDescription>
				</CardHeader>
				<CardContent class="flex justify-center py-6">
					<div class="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
				</CardContent>
			</Card>
		{:else if status === 'success'}
			<Card>
				<CardHeader>
					<CardTitle class="text-center text-green-600 dark:text-green-400">Email Updated!</CardTitle>
					<CardDescription class="text-center">Your email address has been successfully changed.</CardDescription>
				</CardHeader>
				<CardContent class="space-y-4 text-center">
					<p class="text-sm text-muted-foreground">Redirecting you to your profile...</p>
					<div class="flex justify-center">
						<div class="h-8 w-8 animate-spin rounded-full border-4 border-green-500 border-t-transparent"></div>
					</div>
				</CardContent>
			</Card>
		{:else if status === 'error'}
			<Card>
				<CardHeader>
					<CardTitle class="text-center text-destructive">Update Failed</CardTitle>
					<CardDescription class="text-center">We couldn't update your email address.</CardDescription>
				</CardHeader>
				<CardContent class="space-y-4">
					<div class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
						{errorMessage}
					</div>

					<p class="text-center text-sm text-muted-foreground">
						<a href="/account" class="text-primary hover:underline">Back to settings</a>
					</p>
				</CardContent>
			</Card>
		{/if}
	</div>
</div>