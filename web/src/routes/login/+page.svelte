<script lang="ts">
	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Eye, EyeOff } from '@lucide/svelte';
	import { goto } from '$app/navigation';

	let email = $state('');
	let password = $state('');
	let rememberMe = $state(false);
	let showPassword = $state(false);
	let errors = $state<Record<string, string>>({});
	let isLoading = $state(false);
	let generalError = $state('');

	const showResetSuccess = $derived($page.url.searchParams.get('reset') === 'success');

	if (typeof window !== 'undefined' && localStorage.getItem('access_token')) {
		goto('/');
	}

	function validateForm(): boolean {
		const newErrors: Record<string, string> = {};
		if (!email.trim()) {
			newErrors.email = 'Email is required';
		} else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
			newErrors.email = 'Please enter a valid email';
		}
		if (!password) {
			newErrors.password = 'Password is required';
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
			const response = await fetch('/api/auth/login', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ email, password }),
			});

			if (!response.ok) {
				const data = await response.json();
				generalError = data.detail || 'Sign in failed. Please try again.';
				return;
			}

			const data = await response.json();
			localStorage.setItem('access_token', data.access_token);
			localStorage.setItem('refresh_token', data.refresh_token);
			if (rememberMe) localStorage.setItem('remember_me', 'true');
			goto('/');
		} catch {
			generalError = 'Network error. Please try again.';
		} finally {
			isLoading = false;
		}
	}

	function handleForgotPassword() {
		goto('/forgot-password');
	}
</script>

<div class="flex min-h-screen flex-col justify-center bg-background px-4">
	<div class="mx-auto w-full max-w-sm">
		<div class="mb-8 text-center">
			<h1 class="text-2xl font-bold tracking-tight">Welcome back</h1>
			<p class="text-muted-foreground text-sm mt-1">Sign in to manage your medications</p>
		</div>

		{#if generalError}
			<div class="mb-4 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
				{generalError}
			</div>
		{/if}

		{#if showResetSuccess}
			<div class="mb-4 rounded-md bg-green-500/10 p-3 text-sm text-green-600 dark:text-green-400">
				Password reset successful. You can now sign in with your new password.
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
					autocomplete="username"
					aria-describedby={errors.email ? 'email-error' : undefined}
					disabled={isLoading}
				/>
			{#if errors.email}
				<p id="email-error" class="text-sm text-destructive">{errors.email}</p>
				{/if}
			</div>

			<div class="space-y-2">
				<Label for="password">Password</Label>
				<div class="relative">
					<Input
						id="password"
						type={showPassword ? 'text' : 'password'}
						placeholder="Your password"
						bind:value={password}
						class="pr-10"
						autocomplete="current-password"
						aria-describedby={errors.password ? 'password-error' : undefined}
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
				{#if errors.password}
					<p id="password-error" class="text-sm text-destructive">{errors.password}</p>
				{/if}
			</div>

			<div class="flex items-center justify-between">
				<label for="rememberMe" class="flex items-center gap-2 text-sm text-muted-foreground">
					<input id="rememberMe" type="checkbox" bind:checked={rememberMe} class="h-4 w-4 rounded border-input accent-primary" />
					Remember me
				</label>
				<button type="button" onclick={handleForgotPassword} class="text-sm text-primary hover:underline">Forgot password?</button>
			</div>

			<Button type="submit" class="w-full" disabled={isLoading}>
				{#if isLoading}
					Signing in...
				{:else}
					Continue
				{/if}
			</Button>

			<div class="relative my-4">
				<div class="absolute inset-0 flex items-center">
					<span class="w-full border-t border-border"></span>
				</div>
				<div class="relative flex justify-center text-xs uppercase">
					<span class="bg-background px-2 text-muted-foreground">or</span>
				</div>
			</div>

			<div class="grid grid-cols-1 gap-3">
				<Button variant="outline" type="button" onclick={() => alert('Coming soon')}>
					<svg class="mr-2 h-4 w-4" viewBox="0 0 24 24"><path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/><path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/><path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.84z" fill="#FBBC05"/><path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/></svg>
					Google
				</Button>
			</div>
		</form>

		<p class="mt-6 text-center text-sm text-muted-foreground">
			Don't have an account?
			<a href="/register" class="text-primary hover:underline">Sign up</a>
		</p>
	</div>
</div>