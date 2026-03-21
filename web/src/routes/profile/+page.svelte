<script lang="ts">
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import UserRound from '@lucide/svelte/icons/user-round';
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
</div>
