<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { goto } from '$app/navigation';

	let email = $state('');
	let errors = $state<Record<string, string>>({});
	let isLoading = $state(false);
	let generalError = $state('');
	let successMessage = $state('');

	function validateForm(): boolean {
		const newErrors: Record<string, string> = {};
		if (!email.trim()) {
			newErrors.email = 'Email is required';
		} else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
			newErrors.email = 'Please enter a valid email';
		}
		errors = newErrors;
		return Object.keys(newErrors).length === 0;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		generalError = '';
		successMessage = '';
		if (!validateForm()) return;

		isLoading = true;
		try {
			const response = await fetch('/api/auth/password/reset/request', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email }),
			});

			if (!response.ok) {
				const data = await response.json();
				generalError = data.detail || 'Failed to send reset link. Please try again.';
				return;
			}

			// Always show success message for security (don't reveal if email exists)
			successMessage = 'If the email exists, a reset link has been sent';
		} catch {
			generalError = 'Network error. Please try again.';
		} finally {
			isLoading = false;
		}
	}
</script>

<div class="flex min-h-screen flex-col items-center justify-center px-4">
	<div class="mb-8 text-center">
		<h1 class="text-3xl font-bold tracking-tight">MedMinder</h1>
		<p class="text-muted-foreground text-sm mt-1">Your personal medication reminder</p>
	</div>

	<div class="w-full max-w-sm">
		<div class="mb-6 text-center">
			<h2 class="text-xl font-semibold tracking-tight">Forgot Password</h2>
			<p class="text-muted-foreground text-sm mt-1">Enter your email to receive a reset link</p>
		</div>

		{#if generalError}
			<div class="mb-4 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
				{generalError}
			</div>
		{/if}

		{#if successMessage}
			<div class="mb-4 rounded-md bg-green-500/10 p-3 text-sm text-green-600 dark:text-green-400">
				{successMessage}
			</div>
		{/if}

		<form onsubmit={handleSubmit} class="space-y-5">
			<div class="space-y-2">
				<Label for="email">Email</Label>
				<Input
					id="email"
					type="email"
					placeholder="you@example.com"
					bind:value={email}
					variant={errors.email ? 'error' : 'default'}
					autocomplete="email"
					aria-describedby={errors.email ? 'email-error' : undefined}
					disabled={isLoading || !!successMessage}
				/>
				{#if errors.email}
					<p id="email-error" class="text-sm text-destructive">{errors.email}</p>
				{/if}
			</div>

			<Button type="submit" class="w-full" disabled={isLoading || !!successMessage}>
				{#if isLoading}
					Sending...
				{:else}
					Send Reset Link
				{/if}
			</Button>
		</form>

		<p class="mt-6 text-center text-sm text-muted-foreground">
			Remember your password? <a href="/login" class="text-primary hover:underline">Sign in</a>
		</p>
	</div>
</div>