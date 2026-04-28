<script lang="ts">
	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Eye, EyeOff } from '@lucide/svelte';
	import { goto } from '$app/navigation';

	let token = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let showPassword = $state(false);
	let errors = $state<Record<string, string>>({});
	let isLoading = $state(false);
	let generalError = $state('');

	// Get token from URL on mount
	$effect(() => {
		const urlToken = $page.url.searchParams.get('token');
		if (urlToken) {
			token = urlToken;
		} else {
			// Redirect to login if no token
			goto('/login');
		}
	});

	function validateForm(): boolean {
		const newErrors: Record<string, string> = {};
		
		if (!token) {
			newErrors.token = 'Invalid reset token';
		}
		if (!newPassword) {
			newErrors.newPassword = 'Password is required';
	} else if (newPassword.length < 8) {
			newErrors.newPassword = 'Password must be at least 8 characters';
		} else if (!/[A-Z]/.test(newPassword)) {
			newErrors.newPassword = 'Password must contain at least 1 uppercase letter';
		} else if (!/[a-z]/.test(newPassword)) {
			newErrors.newPassword = 'Password must contain at least 1 lowercase letter';
		} else if (!/[0-9]/.test(newPassword)) {
			newErrors.newPassword = 'Password must contain at least 1 number';
		}
		if (!confirmPassword) {
			newErrors.confirmPassword = 'Please confirm your password';
		} else if (newPassword !== confirmPassword) {
			newErrors.confirmPassword = 'Passwords do not match';
		}

		errors = newErrors;
		return Object.keys(newErrors).length === 0;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		generalError = '';
		if (!validateForm()) return;

		isLoading = true;
		try {
			const response = await fetch('/api/auth/password/reset/confirm', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ token, new_password: newPassword }),
			});

			if (!response.ok) {
				const data = await response.json();
				generalError = data.detail || 'Failed to reset password. Please try again.';
				return;
			}

			goto('/login?reset=success');
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
			<h2 class="text-xl font-semibold tracking-tight">Reset Password</h2>
			<p class="text-muted-foreground text-sm mt-1">Enter your new password</p>
		</div>

		{#if generalError}
			<div class="mb-4 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
				{generalError}
			</div>
		{/if}

		<form onsubmit={handleSubmit} class="space-y-5">
			<div class="space-y-2">
				<Label for="newPassword">New Password</Label>
				<div class="relative">
					<Input
						id="newPassword"
						type={showPassword ? 'text' : 'password'}
						placeholder="At least 8 characters"
						bind:value={newPassword}
						variant={errors.newPassword ? 'error' : 'default'}
						class="pr-10"
						autocomplete="new-password"
						aria-describedby={errors.newPassword ? 'newPassword-error' : undefined}
						disabled={isLoading}
					/>
					<button
						type="button"
						onclick={() => (showPassword = !showPassword)}
						class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
						aria-label={showPassword ? 'Hide password' : 'Show password'}
					>
						{#if showPassword}
							<EyeOff class="h-4 w-4" />
						{:else}
							<Eye class="h-4 w-4" />
						{/if}
					</button>
				</div>
				{#if errors.newPassword}
					<p id="newPassword-error" class="text-sm text-destructive">{errors.newPassword}</p>
				{/if}
			</div>

			<div class="space-y-2">
				<Label for="confirmPassword">Confirm Password</Label>
				<Input
					id="confirmPassword"
					type={showPassword ? 'text' : 'password'}
					placeholder="Confirm your password"
					bind:value={confirmPassword}
					variant={errors.confirmPassword ? 'error' : 'default'}
					autocomplete="new-password"
					aria-describedby={errors.confirmPassword ? 'confirmPassword-error' : undefined}
					disabled={isLoading}
				/>
				{#if errors.confirmPassword}
					<p id="confirmPassword-error" class="text-sm text-destructive">{errors.confirmPassword}</p>
				{/if}
			</div>

			<Button type="submit" class="w-full" disabled={isLoading}>
				{#if isLoading}
					Resetting...
				{:else}
					Reset Password
				{/if}
			</Button>
		</form>

		<p class="mt-6 text-center text-sm text-muted-foreground">
			Remember your password? <a href="/login" class="text-primary hover:underline">Sign in</a>
		</p>
	</div>
</div>