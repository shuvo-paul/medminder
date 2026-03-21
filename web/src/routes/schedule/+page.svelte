<script lang="ts">
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import CalendarClock from '@lucide/svelte/icons/calendar-clock';
</script>

<div class="px-4 py-6">
	<header class="mb-6">
		<h1 class="text-2xl font-semibold tracking-tight">Schedule</h1>
		<p class="text-sm text-muted-foreground">Your upcoming reminders</p>
	</header>

	<Card>
		<CardContent class="flex flex-col items-center gap-4 py-12 text-center">
			<div class="rounded-full bg-muted p-4">
				<CalendarClock class="size-8 text-muted-foreground" />
			</div>
			<div>
				<CardTitle class="text-base">No reminders scheduled</CardTitle>
				<CardDescription class="mt-1">Add medications to create a reminder schedule.</CardDescription>
			</div>
			<Button href="/medications" variant="outline">Go to Medications</Button>
		</CardContent>
	</Card>
</div>
