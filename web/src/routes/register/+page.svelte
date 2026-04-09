<script lang="ts">
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { goto } from '$app/navigation';

	let displayName = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let errors = $state<Record<string, string>>({});
	let isLoading = $state(false);
	let generalError = $state('');

	function validateForm(): boolean {
		const newErrors: Record<string, string> = {};
		
		if (!displayName.trim()) {
			newErrors.displayName = 'Display name is required';
		}
		if (!email.trim()) {
			newErrors.email = 'Email is required';
		} else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
			newErrors.email = 'Please enter a valid email';
		}
		if (!password) {
			newErrors.password = 'Password is required';
		} else if (password.length < 8) {
			newErrors.password = 'Password must be at least 8 characters';
		}
		if (password !== confirmPassword) {
			newErrors.confirmPassword = 'Passwords do not match';
		}

		errors = newErrors;
		return Object.keys(newErrors).length === 0;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		generalError = '';
		
		if (!validateForm()) {
			return;
		}

		isLoading = true;

		try {
			const response = await fetch('/api/auth/register', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({
					email,
					password,
					display_name: displayName,
				}),
			});

			const data = await response.json();

			if (!response.ok) {
				if (data.detail?.includes('email')) {
					errors.email = 'Email already exists';
				} else {
					generalError = data.detail || 'Registration failed. Please try again.';
				}
				return;
			}

			localStorage.setItem('access_token', data.access_token);
			localStorage.setItem('refresh_token', data.refresh_token);
			goto('/');
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

	<Card class="w-full max-w-sm">
		<CardHeader>
			<CardTitle>Create Account</CardTitle>
			<CardDescription>Enter your details to get started</CardDescription>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSubmit} class="space-y-4">
				{#if generalError}
					<div class="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
						{generalError}
					</div>
				{/if}

				<div class="space-y-2">
					<Label for="displayName">Display Name</Label>
					<Input
						id="displayName"
						type="text"
						placeholder="Your name"
						bind:value={displayName}
						variant={errors.displayName ? 'error' : 'default'}
						autocomplete="name"
					/>
					{#if errors.displayName}
						<p class="text-sm text-destructive">{errors.displayName}</p>
					{/if}
				</div>

				<div class="space-y-2">
					<Label for="email">Email</Label>
					<Input
						id="email"
						type="email"
						placeholder="you@example.com"
						bind:value={email}
						variant={errors.email ? 'error' : 'default'}
						autocomplete="email"
					/>
					{#if errors.email}
						<p class="text-sm text-destructive">{errors.email}</p>
					{/if}
				</div>

				<div class="space-y-2">
					<Label for="password">Password</Label>
					<Input
						id="password"
						type="password"
						placeholder="At least 8 characters"
						bind:value={password}
						variant={errors.password ? 'error' : 'default'}
						autocomplete="new-password"
					/>
					{#if errors.password}
						<p class="text-sm text-destructive">{errors.password}</p>
					{/if}
				</div>

				<div class="space-y-2">
					<Label for="confirmPassword">Confirm Password</Label>
					<Input
						id="confirmPassword"
						type="password"
						placeholder="Confirm your password"
						bind:value={confirmPassword}
						variant={errors.confirmPassword ? 'error' : 'default'}
						autocomplete="new-password"
					/>
					{#if errors.confirmPassword}
						<p class="text-sm text-destructive">{errors.confirmPassword}</p>
					{/if}
				</div>

				<Button type="submit" class="w-full" disabled={isLoading}>
					{#if isLoading}
						Creating account...
					{:else}
						Sign Up
					{/if}
				</Button>

				<p class="text-center text-sm text-muted-foreground">
					Already have an account?
					<a href="/login" class="text-primary hover:underline">Sign in</a>
				</p>
			</form>
		</CardContent>
	</Card>
</div>