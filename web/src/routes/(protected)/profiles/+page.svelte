<script lang="ts">
	import { onMount } from 'svelte';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { toast } from 'svelte-sonner';
	import UsersRound from '@lucide/svelte/icons/users-round';
	import Plus from '@lucide/svelte/icons/plus';
	import Clock from '@lucide/svelte/icons/clock';
	import Globe from '@lucide/svelte/icons/globe';
	import { listProfiles, type ProfileDTO } from '$lib/api/profiles';

	let profiles = $state<ProfileDTO[]>([]);
	let isLoading = $state(true);
	let error = $state('');

	async function loadProfiles() {
		isLoading = true;
		error = '';
		try {
			profiles = await listProfiles();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load profiles';
			toast.error(error);
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		loadProfiles();
	});
</script>

<div class="px-4 py-6">
	<header class="mb-6 flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Profiles</h1>
			<p class="text-sm text-muted-foreground">Manage medication profiles for you and your family</p>
		</div>
		<Button href="/profiles/new">
			<Plus class="size-4" />
			New Profile
		</Button>
	</header>

	{#if isLoading}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each Array(3) as _}
				<Card class="animate-pulse">
					<CardHeader>
						<div class="h-5 w-24 rounded bg-muted"></div>
						<div class="h-4 w-32 rounded bg-muted"></div>
					</CardHeader>
					<CardContent>
						<div class="h-4 w-40 rounded bg-muted"></div>
					</CardContent>
				</Card>
			{/each}
		</div>
	{:else if error && profiles.length === 0}
		<Card>
			<CardContent class="flex flex-col items-center gap-4 py-12 text-center">
				<div class="rounded-full bg-muted p-4">
					<UsersRound class="size-8 text-muted-foreground" />
				</div>
				<div>
					<CardTitle class="text-base">Failed to load profiles</CardTitle>
					<CardDescription class="mt-1">{error}</CardDescription>
				</div>
				<Button variant="outline" onclick={loadProfiles}>Try Again</Button>
			</CardContent>
		</Card>
	{:else if profiles.length === 0}
		<Card>
			<CardContent class="flex flex-col items-center gap-4 py-12 text-center">
				<div class="rounded-full bg-muted p-4">
					<UsersRound class="size-8 text-muted-foreground" />
				</div>
				<div>
					<CardTitle class="text-base">No profiles yet</CardTitle>
					<CardDescription class="mt-1">Create your first medication profile to get started.</CardDescription>
				</div>
				<Button href="/profiles/new">
					<Plus class="size-4" />
					Create Profile
				</Button>
			</CardContent>
		</Card>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			{#each profiles as profile (profile.id)}
				<a href="/profiles/{profile.id}" class="block">
					<Card class="transition-colors hover:border-primary/50">
						<CardHeader>
							<div class="flex items-start justify-between">
								<CardTitle class="text-base">{profile.name}</CardTitle>
								<Badge variant="secondary">{profile.schedules.length} schedule{profile.schedules.length !== 1 ? 's' : ''}</Badge>
							</div>
							{#if profile.date_of_birth}
								<CardDescription>DOB: {profile.date_of_birth}</CardDescription>
							{/if}
						</CardHeader>
						<CardContent class="flex flex-col gap-2">
							<div class="flex items-center gap-1.5 text-sm text-muted-foreground">
								<Globe class="size-3.5" />
								<span>{profile.timezone}</span>
							</div>
							{#if profile.schedules.length > 0}
								<div class="flex items-center gap-1.5 text-sm text-muted-foreground">
									<Clock class="size-3.5" />
									<span>{profile.schedules.map((s) => `${s.name} ${s.time}`).join(' · ')}</span>
								</div>
							{/if}
						</CardContent>
					</Card>
				</a>
			{/each}
		</div>
	{/if}
</div>