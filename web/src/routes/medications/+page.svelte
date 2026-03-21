<script lang="ts">
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import Pill from '@lucide/svelte/icons/pill';
</script>

<div class="px-4 py-6">
	<header class="mb-6">
		<h1 class="text-2xl font-semibold tracking-tight">Medications</h1>
		<p class="text-sm text-muted-foreground">Manage your medication list</p>
	</header>

	<Card>
		<CardContent class="flex flex-col items-center gap-4 py-12 text-center">
			<div class="rounded-full bg-muted p-4">
				<Pill class="size-8 text-muted-foreground" />
			</div>
			<div>
				<CardTitle class="text-base">No medications yet</CardTitle>
				<CardDescription class="mt-1">Add your first medication to get started.</CardDescription>
			</div>
			<Button>Add Medication</Button>
		</CardContent>
	</Card>
</div>
