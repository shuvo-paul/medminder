<script lang="ts">
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { ChevronLeft, Plus, Trash2 } from '@lucide/svelte/icons';
	import { createProfile } from '$lib/api/profiles';

	const COMMON_TIMEZONES = [
		'America/New_York',
		'America/Chicago',
		'America/Denver',
		'America/Los_Angeles',
		'America/Anchorage',
		'Pacific/Honolulu',
		'America/Toronto',
		'America/Vancouver',
		'America/Mexico_City',
		'America/Sao_Paulo',
		'America/Buenos_Aires',
		'Europe/London',
		'Europe/Paris',
		'Europe/Berlin',
		'Europe/Rome',
		'Europe/Madrid',
		'Europe/Amsterdam',
		'Europe/Stockholm',
		'Europe/Moscow',
		'Asia/Dubai',
		'Asia/Kolkata',
		'Asia/Dhaka',
		'Asia/Bangkok',
		'Asia/Singapore',
		'Asia/Hong_Kong',
		'Asia/Shanghai',
		'Asia/Tokyo',
		'Asia/Seoul',
		'Australia/Sydney',
		'Australia/Melbourne',
		'Pacific/Auckland',
		'UTC',
	];

	const defaultTimezone = 'Asia/Dhaka';

	let name = $state('');
	let dateOfBirth = $state('');
	let timezone = $state(defaultTimezone);
	let errors = $state<Record<string, string>>({});
	let showTzDropdown = $state(false);
	let tzFiltered = $derived(
		COMMON_TIMEZONES.filter((tz) => tz.toLowerCase().includes(timezone.toLowerCase()))
	);
	let isSubmitting = $state(false);

	interface ScheduleRow {
		name: string;
		time: string;
	}

	let schedules = $state<ScheduleRow[]>([
		{ name: 'Breakfast', time: '08:00' },
		{ name: 'Lunch', time: '13:00' },
		{ name: 'Dinner', time: '19:00' },
	]);

	function addSchedule() {
		schedules = [...schedules, { name: '', time: '12:00' }];
	}

	function removeSchedule(index: number) {
		schedules = schedules.filter((_, i) => i !== index);
	}

	function validateForm(): boolean {
		errors = {};
		if (!name.trim()) {
			errors.name = 'Name is required';
		} else if (name.trim().length > 100) {
			errors.name = 'Name must be 100 characters or less';
		}
		if (dateOfBirth && !/^\d{4}-\d{2}-\d{2}$/.test(dateOfBirth)) {
			errors.dateOfBirth = 'Date must be in YYYY-MM-DD format';
		}
		if (!timezone) {
			errors.timezone = 'Timezone is required';
		}
		for (let i = 0; i < schedules.length; i++) {
			const s = schedules[i];
			if (!s.name.trim()) {
				errors[`schedule_${i}_name`] = 'Schedule name is required';
			}
			if (!s.time.trim()) {
				errors[`schedule_${i}_time`] = 'Schedule time is required';
			} else if (!/^\d{2}:\d{2}$/.test(s.time)) {
				errors[`schedule_${i}_time`] = 'Time must be in HH:MM format';
			}
		}
		return Object.keys(errors).length === 0;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!validateForm()) return;

		isSubmitting = true;
		try {
			const validSchedules = schedules
				.filter((s) => s.name.trim() && s.time.trim())
				.map((s) => ({ name: s.name.trim(), time: s.time }));

			await createProfile({
				name: name.trim(),
				date_of_birth: dateOfBirth || null,
				timezone,
				dose_schedules: validSchedules.length > 0 ? validSchedules : undefined,
			});
			toast.success('Profile created successfully');
			goto('/profiles');
		} catch (err) {
			toast.error(err instanceof Error ? err.message : 'Failed to create profile');
		} finally {
			isSubmitting = false;
		}
	}
</script>

<div class="px-4 py-6">
	<header class="mb-6">
		<a href="/profiles" class="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
			<ChevronLeft class="size-4" />
			Back to Profiles
		</a>
		<h1 class="text-2xl font-semibold tracking-tight">Create Profile</h1>
		<p class="text-sm text-muted-foreground">Add a new medication profile</p>
	</header>

	<form onsubmit={handleSubmit} class="max-w-lg space-y-6">
		<Card>
			<CardHeader>
				<CardTitle>Profile Details</CardTitle>
				<CardDescription>Basic information for this profile</CardDescription>
			</CardHeader>
			<CardContent class="space-y-4">
				<div class="space-y-2">
					<Label for="name">Name <span class="text-destructive">*</span></Label>
					<Input
						id="name"
						bind:value={name}
						placeholder="e.g., John's Profile"
						class={errors.name ? 'border-destructive' : ''}
						aria-describedby={errors.name ? 'name-error' : undefined}
					/>
					{#if errors.name}
						<p id="name-error" class="text-sm text-destructive">{errors.name}</p>
					{/if}
				</div>

				<div class="space-y-2">
					<Label for="dateOfBirth">Date of Birth</Label>
					<Input
						id="dateOfBirth"
						type="date"
						bind:value={dateOfBirth}
						class={errors.dateOfBirth ? 'border-destructive' : ''}
						aria-describedby={errors.dateOfBirth ? 'dob-error' : undefined}
					/>
					{#if errors.dateOfBirth}
						<p id="dob-error" class="text-sm text-destructive">{errors.dateOfBirth}</p>
					{/if}
				</div>

				<div class="space-y-2">
					<Label for="timezone">Timezone <span class="text-destructive">*</span></Label>
					<div class="relative">
						<Input
							id="timezone"
							placeholder="Search timezone..."
							autocomplete="off"
							bind:value={timezone}
							onfocus={() => (showTzDropdown = true)}
							onblur={() => (showTzDropdown = false)}
						/>
						{#if showTzDropdown}
							<div class="absolute z-50 mt-1 max-h-48 w-full overflow-auto rounded-md border bg-popover shadow-md">
								{#each tzFiltered as tz}
									<button
										type="button"
										class="w-full px-3 py-1.5 text-left text-sm hover:bg-accent hover:text-accent-foreground"
										onmousedown={() => {
											timezone = tz;
											showTzDropdown = false;
										}}
									>{tz}</button>
								{/each}
								{#if tzFiltered.length === 0}
									<div class="px-3 py-1.5 text-sm text-muted-foreground">No timezone found</div>
								{/if}
							</div>
						{/if}
					</div>
					{#if errors.timezone}
						<p class="text-sm text-destructive">{errors.timezone}</p>
					{/if}
				</div>
			</CardContent>
		</Card>

		<Card>
			<CardHeader>
				<CardTitle>Initial Dose Schedules</CardTitle>
				<CardDescription>Set up meal times to create default reminders</CardDescription>
			</CardHeader>
			<CardContent class="space-y-4">
				{#each schedules as schedule, i (i)}
					<div class="flex items-end gap-2">
						<div class="flex-1 space-y-2">
							<Label for="schedule-name-{i}">Name</Label>
							<Input
								id="schedule-name-{i}"
								bind:value={schedule.name}
								placeholder="e.g., Breakfast"
								class={errors[`schedule_${i}_name`] ? 'border-destructive' : ''}
							/>
							{#if errors[`schedule_${i}_name`]}
								<p class="text-sm text-destructive">{errors[`schedule_${i}_name`]}</p>
							{/if}
						</div>
						<div class="w-28 space-y-2">
							<Label for="schedule-time-{i}">Time</Label>
							<Input
								id="schedule-time-{i}"
								type="time"
								bind:value={schedule.time}
								class={errors[`schedule_${i}_time`] ? 'border-destructive' : ''}
							/>
							{#if errors[`schedule_${i}_time`]}
								<p class="text-sm text-destructive">{errors[`schedule_${i}_time`]}</p>
							{/if}
						</div>
						<Button
							type="button"
							variant="ghost"
							size="icon"
							onclick={() => removeSchedule(i)}
							disabled={schedules.length <= 1}
							class="mb-0.5 shrink-0"
						>
							<Trash2 class="size-4 text-muted-foreground" />
						</Button>
					</div>
				{/each}

				<Button type="button" variant="outline" size="sm" onclick={addSchedule}>
					<Plus class="size-4" />
					Add Schedule
				</Button>
			</CardContent>
		</Card>

		{#if errors.general}
			<p class="text-sm text-destructive">{errors.general}</p>
		{/if}

		<div class="flex gap-3">
			<Button type="submit" disabled={isSubmitting}>
				{isSubmitting ? 'Creating...' : 'Create Profile'}
			</Button>
			<Button type="button" variant="outline" href="/profiles" disabled={isSubmitting}>
				Cancel
			</Button>
		</div>
	</form>
</div>