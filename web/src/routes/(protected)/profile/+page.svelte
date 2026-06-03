<script lang="ts">
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import UserRound from '@lucide/svelte/icons/user-round';
	import { goto } from '$app/navigation';

	async function handleLogout() {
		const token = localStorage.getItem('access_token');
		if (token) {
			await fetch('/api/auth/logout', {
				method: 'POST',
				headers: {
					Authorization: `Bearer ${token}`,
				},
			});
		}
		localStorage.removeItem('access_token');
		localStorage.removeItem('refresh_token');
		localStorage.removeItem('email_verified');
		localStorage.removeItem('remember_me');
		goto('/login');
	}
</script>

<div class="px-4 py-6">
	<header class="mb-6">
		<h1 class="text-2xl font-semibold tracking-tight">Profile</h1>
		<p class="text-sm text-muted-foreground">Account &amp; settings</p>
	</header>

	<Card>
		<CardContent class="flex flex-col items-center gap-4 py-12 text-center">
			<div class="rounded-full bg-muted p-4">
				<UserRound class="size-8 text-muted-foreground" />
			</div>
			<div>
				<CardTitle class="text-base">Set up your profile</CardTitle>
				<CardDescription class="mt-1">Personalize your MedMinder experience.</CardDescription>
			</div>
			<Button variant="outline">Edit Profile</Button>
		</CardContent>
	</Card>

	<div class="mt-6">
		<Button variant="ghost" class="w-full text-destructive" onclick={handleLogout}>
			Sign Out
		</Button>
	</div>
</div>
