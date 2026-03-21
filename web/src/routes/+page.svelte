<script lang="ts">
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
</script>

<div class="px-4 py-6">
	<header class="mb-6">
		<p class="text-sm text-muted-foreground">Good morning</p>
		<h1 class="text-2xl font-semibold tracking-tight">Dashboard</h1>
	</header>

	<div class="space-y-4">
		<Card>
			<CardHeader>
				<div class="flex items-start justify-between">
					<CardTitle class="text-base">Today's Schedule</CardTitle>
					<Badge variant="secondary">0 due</Badge>
				</div>
				<CardDescription>No medications scheduled for today.</CardDescription>
			</CardHeader>
			<CardContent>
				<Button href="/medications" variant="outline" class="w-full">Add Medication</Button>
			</CardContent>
		</Card>

		<Card>
			<CardHeader>
				<CardTitle class="text-base">Adherence</CardTitle>
				<CardDescription>Track your medication consistency over time.</CardDescription>
			</CardHeader>
			<CardContent>
				<p class="text-sm text-muted-foreground">No data yet. Start tracking to see your adherence.</p>
			</CardContent>
		</Card>
	</div>
</div>
