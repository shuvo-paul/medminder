<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Badge } from '$lib/components/ui/badge';
	import {
		Dialog,
		DialogContent,
		DialogDescription,
		DialogFooter,
		DialogHeader,
		DialogTitle,
	} from '$lib/components/ui/dialog';
	import {
		ChevronLeft,
		Pencil,
		Trash2,
		Plus,
		Clock,
		Globe,
		Calendar,
	} from '@lucide/svelte/icons';
	import {
		getProfile,
		updateProfile,
		deleteProfile,
		createSchedule,
		updateSchedule,
		deleteSchedule,
		listSchedules,
		type ProfileDTO,
		type DoseScheduleDTO,
	} from '$lib/api/profiles';

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

	const profileId = $derived($page.params.id as string);

	let profile = $state<ProfileDTO | null>(null);
	let schedules = $state<DoseScheduleDTO[]>([]);
	let isLoading = $state(true);
	let loadError = $state('');
	let notFound = $state(false);

	let isEditing = $state(false);
	let editName = $state('');
	let editDateOfBirth = $state('');
	let editTimezone = $state('');
	let editErrors = $state<Record<string, string>>({});
	let showTzDropdown = $state(false);
	let tzHideTimer: ReturnType<typeof setTimeout>;
	let tzFiltered = $derived(
		COMMON_TIMEZONES.filter((tz) => tz.toLowerCase().includes(editTimezone.toLowerCase()))
	);
	let isSavingProfile = $state(false);

	let showAddSchedule = $state(false);
	let newScheduleName = $state('');
	let newScheduleTime = $state('12:00');
	let addScheduleErrors = $state<Record<string, string>>({});
	let isAddingSchedule = $state(false);

	let editingScheduleId = $state<string | null>(null);
	let editScheduleName = $state('');
	let editScheduleTime = $state('');
	let editScheduleErrors = $state<Record<string, string>>({});
	let isUpdatingSchedule = $state(false);

	let deletingScheduleId = $state<string | null>(null);
	let isDeletingSchedule = $state(false);

	let showDeleteProfile = $state(false);
	let isDeletingProfile = $state(false);

	async function loadData() {
		isLoading = true;
		loadError = '';
		notFound = false;
		try {
			profile = await getProfile(profileId);
			schedules = profile.schedules;
			editName = profile.name;
			editDateOfBirth = profile.date_of_birth || '';
			editTimezone = profile.timezone;
		} catch (e) {
			if (e instanceof Error && e.message.includes('404')) {
				notFound = true;
			} else {
				loadError = e instanceof Error ? e.message : 'Failed to load profile';
			}
		} finally {
			isLoading = false;
		}
	}

	onMount(() => {
		loadData();
	});

	function startEdit() {
		if (profile) {
			editName = profile.name;
			editDateOfBirth = profile.date_of_birth || '';
			editTimezone = profile.timezone;
			isEditing = true;
		}
	}

	function cancelEdit() {
		isEditing = false;
		editErrors = {};
	}

	function validateEditForm(): boolean {
		editErrors = {};
		if (!editName.trim()) {
			editErrors.name = 'Name is required';
		} else if (editName.trim().length > 100) {
			editErrors.name = 'Name must be 100 characters or less';
		}
		if (editDateOfBirth && !/^\d{4}-\d{2}-\d{2}$/.test(editDateOfBirth)) {
			editErrors.dateOfBirth = 'Date must be in YYYY-MM-DD format';
		}
		return Object.keys(editErrors).length === 0;
	}

	async function saveProfile() {
		if (!validateEditForm()) return;
		isSavingProfile = true;
		try {
			profile = await updateProfile(profileId, {
				name: editName.trim(),
				date_of_birth: editDateOfBirth || null,
				timezone: editTimezone,
			});
			isEditing = false;
			toast.success('Profile updated');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to update profile');
		} finally {
			isSavingProfile = false;
		}
	}

	async function confirmDeleteProfile() {
		isDeletingProfile = true;
		try {
			await deleteProfile(profileId);
			toast.success('Profile deleted');
			goto('/profiles');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to delete profile');
		} finally {
			isDeletingProfile = false;
		}
	}

	function openAddSchedule() {
		newScheduleName = '';
		newScheduleTime = '12:00';
		addScheduleErrors = {};
		showAddSchedule = true;
	}

	function validateAddSchedule(): boolean {
		addScheduleErrors = {};
		if (!newScheduleName.trim()) {
			addScheduleErrors.name = 'Name is required';
		}
		if (!newScheduleTime.trim()) {
			addScheduleErrors.time = 'Time is required';
		} else if (!/^\d{2}:\d{2}$/.test(newScheduleTime)) {
			addScheduleErrors.time = 'Time must be in HH:MM format';
		}
		return Object.keys(addScheduleErrors).length === 0;
	}

	async function handleAddSchedule() {
		if (!validateAddSchedule()) return;
		isAddingSchedule = true;
		try {
			const newSched = await createSchedule(profileId, {
				name: newScheduleName.trim(),
				time: newScheduleTime,
			});
			schedules = [...schedules, newSched];
			showAddSchedule = false;
			toast.success('Schedule added');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to add schedule');
		} finally {
			isAddingSchedule = false;
		}
	}

	function openEditSchedule(sched: DoseScheduleDTO) {
		editingScheduleId = sched.id;
		editScheduleName = sched.name;
		editScheduleTime = sched.time;
		editScheduleErrors = {};
	}

	function closeEditSchedule() {
		editingScheduleId = null;
		editScheduleErrors = {};
	}

	function validateEditSchedule(): boolean {
		editScheduleErrors = {};
		if (!editScheduleName.trim()) {
			editScheduleErrors.name = 'Name is required';
		}
		if (!editScheduleTime.trim()) {
			editScheduleErrors.time = 'Time is required';
		} else if (!/^\d{2}:\d{2}$/.test(editScheduleTime)) {
			editScheduleErrors.time = 'Time must be in HH:MM format';
		}
		return Object.keys(editScheduleErrors).length === 0;
	}

	async function handleUpdateSchedule() {
		if (!validateEditSchedule()) return;
		isUpdatingSchedule = true;
		try {
			const updated = await updateSchedule(profileId, editingScheduleId!, {
				name: editScheduleName.trim(),
				time: editScheduleTime,
			});
			schedules = schedules.map((s) => (s.id === editingScheduleId ? updated : s));
			closeEditSchedule();
			toast.success('Schedule updated');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to update schedule');
		} finally {
			isUpdatingSchedule = false;
		}
	}

	function openDeleteSchedule(sched: DoseScheduleDTO) {
		deletingScheduleId = sched.id;
	}

	function closeDeleteSchedule() {
		deletingScheduleId = null;
	}

	async function handleDeleteSchedule() {
		if (!deletingScheduleId) return;
		isDeletingSchedule = true;
		try {
			await deleteSchedule(profileId, deletingScheduleId);
			schedules = schedules.filter((s) => s.id !== deletingScheduleId);
			closeDeleteSchedule();
			toast.success('Schedule deleted');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to delete schedule');
		} finally {
			isDeletingSchedule = false;
		}
	}
</script>

<div class="px-4 py-6">
	<a href="/profiles" class="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
		<ChevronLeft class="size-4" />
		Back to Profiles
	</a>

	{#if isLoading}
		<div class="space-y-4">
			<div class="h-8 w-48 animate-pulse rounded bg-muted"></div>
			<div class="h-4 w-72 animate-pulse rounded bg-muted"></div>
			<div class="mt-6 space-y-4">
				{#each Array(2) as _}
					<Card class="animate-pulse">
						<CardHeader><div class="h-5 w-32 rounded bg-muted"></div></CardHeader>
						<CardContent><div class="h-4 w-48 rounded bg-muted"></div></CardContent>
					</Card>
				{/each}
			</div>
		</div>
	{:else if notFound}
		<Card>
			<CardContent class="flex flex-col items-center gap-4 py-12 text-center">
				<CardTitle class="text-base">Profile not found</CardTitle>
				<CardDescription>This profile may have been deleted or you don't have access.</CardDescription>
				<Button href="/profiles" variant="outline">Back to Profiles</Button>
			</CardContent>
		</Card>
	{:else if loadError}
		<Card>
			<CardContent class="flex flex-col items-center gap-4 py-12 text-center">
				<CardTitle class="text-base">Failed to load profile</CardTitle>
				<CardDescription>{loadError}</CardDescription>
				<Button variant="outline" onclick={loadData}>Try Again</Button>
			</CardContent>
		</Card>
	{:else if profile}
		<header class="mb-6 flex items-start justify-between">
			<div>
				<h1 class="text-2xl font-semibold tracking-tight">{profile.name}</h1>
				<p class="text-sm text-muted-foreground">Profile details and schedules</p>
			</div>
			{#if !isEditing}
				<Button variant="outline" size="sm" onclick={startEdit}>
					<Pencil class="size-4" />
					Edit Profile
				</Button>
			{/if}
		</header>

		<div class="space-y-6">
			<Card>
				<CardHeader>
					<div class="flex items-center justify-between">
						<CardTitle>Profile Information</CardTitle>
						{#if isEditing}
							<div class="flex gap-2">
								<Button size="sm" variant="outline" onclick={cancelEdit} disabled={isSavingProfile}>Cancel</Button>
								<Button size="sm" onclick={saveProfile} disabled={isSavingProfile}>
									{isSavingProfile ? 'Saving...' : 'Save'}
								</Button>
							</div>
						{/if}
					</div>
				</CardHeader>
				<CardContent>
					{#if isEditing}
						<div class="space-y-4">
							<div class="space-y-2">
								<Label for="edit-name">Name</Label>
								<Input
									id="edit-name"
									bind:value={editName}
									class={editErrors.name ? 'border-destructive' : ''}
								/>
								{#if editErrors.name}
									<p class="text-sm text-destructive">{editErrors.name}</p>
								{/if}
							</div>
							<div class="space-y-2">
								<Label for="edit-dob">Date of Birth</Label>
								<Input
									id="edit-dob"
									type="date"
									bind:value={editDateOfBirth}
									class={editErrors.dateOfBirth ? 'border-destructive' : ''}
								/>
								{#if editErrors.dateOfBirth}
									<p class="text-sm text-destructive">{editErrors.dateOfBirth}</p>
								{/if}
							</div>
							<div class="space-y-2">
								<Label for="edit-tz">Timezone</Label>
								<div class="relative">
									<Input
										id="edit-tz"
										placeholder="Search timezone..."
										autocomplete="off"
										bind:value={editTimezone}
										onfocus={() => {
											clearTimeout(tzHideTimer);
											showTzDropdown = true;
										}}
										onblur={() => {
											tzHideTimer = setTimeout(() => {
												showTzDropdown = false;
											}, 150);
										}}
									/>
									{#if showTzDropdown}
										<div
											role="listbox"
											class="absolute z-50 mt-1 max-h-48 w-full overflow-auto rounded-md border bg-popover shadow-md"
											onmouseenter={() => clearTimeout(tzHideTimer)}
											onmouseleave={() => {
												showTzDropdown = false;
											}}
										>
											{#each tzFiltered as tz}
												<button
													type="button"
													role="option"
													class="w-full px-3 py-1.5 text-left text-sm hover:bg-accent hover:text-accent-foreground"
													onmousedown={() => {
														editTimezone = tz;
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
								{#if editErrors.timezone}
									<p class="text-sm text-destructive">{editErrors.timezone}</p>
								{/if}
							</div>
						</div>
					{:else}
						<dl class="flex flex-col gap-4">
							<div class="flex items-center gap-2">
								<Globe class="size-4 text-muted-foreground" />
								<dt class="text-sm text-muted-foreground">Timezone</dt>
								<dd class="text-sm font-medium">{profile.timezone}</dd>
							</div>
							{#if profile.date_of_birth}
								<div class="flex items-center gap-2">
									<Calendar class="size-4 text-muted-foreground" />
									<dt class="text-sm text-muted-foreground">Date of Birth</dt>
									<dd class="text-sm font-medium">{profile.date_of_birth}</dd>
								</div>
							{/if}
						</dl>
					{/if}
				</CardContent>
			</Card>

			<Card>
				<CardHeader>
					<div class="flex items-center justify-between">
						<CardTitle>Dose Schedules</CardTitle>
						<Button size="sm" variant="outline" onclick={openAddSchedule}>
							<Plus class="size-4" />
							Add Schedule
						</Button>
					</div>
					<CardDescription>Times when medications should be taken</CardDescription>
				</CardHeader>
				<CardContent>
					{#if schedules.length === 0}
						<p class="text-sm text-muted-foreground">No schedules yet. Add one to get started.</p>
					{:else}
						<ul class="divide-y divide-border">
							{#each schedules as sched (sched.id)}
								<li class="flex items-center justify-between py-3">
									<div class="flex items-center gap-3">
										<Clock class="size-4 text-muted-foreground" />
										<div>
											<p class="text-sm font-medium">{sched.name}</p>
											<p class="text-xs text-muted-foreground">{sched.time}</p>
										</div>
									</div>
									<div class="flex gap-1">
										<Button
											variant="ghost"
											size="icon"
											onclick={() => openEditSchedule(sched)}
										>
											<Pencil class="size-4" />
										</Button>
										<Button
											variant="ghost"
											size="icon"
											onclick={() => openDeleteSchedule(sched)}
										>
											<Trash2 class="size-4 text-destructive" />
										</Button>
									</div>
								</li>
							{/each}
						</ul>
					{/if}
				</CardContent>
			</Card>

			<Card class="border-destructive/50">
				<CardHeader>
					<CardTitle class="text-destructive">Danger Zone</CardTitle>
				</CardHeader>
				<CardContent>
					<p class="text-sm text-muted-foreground mb-3">
						Deleting this profile will also delete all associated dose schedules. This action cannot be undone.
					</p>
					<Button variant="destructive" size="sm" onclick={() => (showDeleteProfile = true)}>
						<Trash2 class="size-4" />
						Delete Profile
					</Button>
				</CardContent>
			</Card>
		</div>
	{/if}
</div>

<Dialog open={showAddSchedule} onclose={() => (showAddSchedule = false)}>
	<DialogContent>
		<DialogHeader>
			<DialogTitle>Add Schedule</DialogTitle>
			<DialogDescription>Create a new dose schedule for this profile.</DialogDescription>
		</DialogHeader>
		<div class="space-y-4">
			<div class="space-y-2">
				<Label for="new-sched-name">Name</Label>
				<Input
					id="new-sched-name"
					bind:value={newScheduleName}
					placeholder="e.g., Breakfast"
					class={addScheduleErrors.name ? 'border-destructive' : ''}
				/>
				{#if addScheduleErrors.name}
					<p class="text-sm text-destructive">{addScheduleErrors.name}</p>
				{/if}
			</div>
			<div class="space-y-2">
				<Label for="new-sched-time">Time</Label>
				<Input
					id="new-sched-time"
					type="time"
					bind:value={newScheduleTime}
					class={addScheduleErrors.time ? 'border-destructive' : ''}
				/>
				{#if addScheduleErrors.time}
					<p class="text-sm text-destructive">{addScheduleErrors.time}</p>
				{/if}
			</div>
		</div>
		<DialogFooter>
			<Button variant="outline" onclick={() => (showAddSchedule = false)} disabled={isAddingSchedule}>Cancel</Button>
			<Button onclick={handleAddSchedule} disabled={isAddingSchedule}>
				{isAddingSchedule ? 'Adding...' : 'Add Schedule'}
			</Button>
		</DialogFooter>
	</DialogContent>
</Dialog>

<Dialog open={editingScheduleId !== null} onclose={closeEditSchedule}>
	<DialogContent>
		<DialogHeader>
			<DialogTitle>Edit Schedule</DialogTitle>
		</DialogHeader>
		<div class="space-y-4">
			<div class="space-y-2">
				<Label for="edit-sched-name">Name</Label>
				<Input
					id="edit-sched-name"
					bind:value={editScheduleName}
					class={editScheduleErrors.name ? 'border-destructive' : ''}
				/>
				{#if editScheduleErrors.name}
					<p class="text-sm text-destructive">{editScheduleErrors.name}</p>
				{/if}
			</div>
			<div class="space-y-2">
				<Label for="edit-sched-time">Time</Label>
				<Input
					id="edit-sched-time"
					type="time"
					bind:value={editScheduleTime}
					class={editScheduleErrors.time ? 'border-destructive' : ''}
				/>
				{#if editScheduleErrors.time}
					<p class="text-sm text-destructive">{editScheduleErrors.time}</p>
				{/if}
			</div>
		</div>
		<DialogFooter>
			<Button variant="outline" onclick={closeEditSchedule} disabled={isUpdatingSchedule}>Cancel</Button>
			<Button onclick={handleUpdateSchedule} disabled={isUpdatingSchedule}>
				{isUpdatingSchedule ? 'Saving...' : 'Save Changes'}
			</Button>
		</DialogFooter>
	</DialogContent>
</Dialog>

<Dialog open={deletingScheduleId !== null} onclose={closeDeleteSchedule}>
	<DialogContent>
		<DialogHeader>
			<DialogTitle>Delete Schedule</DialogTitle>
			<DialogDescription>Are you sure you want to delete this schedule? This action cannot be undone.</DialogDescription>
		</DialogHeader>
		<DialogFooter>
			<Button variant="outline" onclick={closeDeleteSchedule} disabled={isDeletingSchedule}>Cancel</Button>
			<Button variant="destructive" onclick={handleDeleteSchedule} disabled={isDeletingSchedule}>
				{isDeletingSchedule ? 'Deleting...' : 'Delete Schedule'}
			</Button>
		</DialogFooter>
	</DialogContent>
</Dialog>

<Dialog open={showDeleteProfile} onclose={() => (showDeleteProfile = false)}>
	<DialogContent>
		<DialogHeader>
			<DialogTitle>Delete Profile</DialogTitle>
			<DialogDescription>
				Are you sure you want to delete <strong>{profile?.name}</strong>? All dose schedules will be permanently deleted. This action cannot be undone.
			</DialogDescription>
		</DialogHeader>
		<DialogFooter>
			<Button variant="outline" onclick={() => (showDeleteProfile = false)} disabled={isDeletingProfile}>Cancel</Button>
			<Button variant="destructive" onclick={confirmDeleteProfile} disabled={isDeletingProfile}>
				{isDeletingProfile ? 'Deleting...' : 'Delete Profile'}
			</Button>
		</DialogFooter>
	</DialogContent>
</Dialog>