function getToken(): string {
	return localStorage.getItem('access_token') || '';
}

function authHeaders(): Record<string, string> {
	return {
		Authorization: `Bearer ${getToken()}`,
	};
}

async function handleResponse<T>(res: Response): Promise<T> {
	if (!res.ok) {
		const err = await res.json().catch(() => ({ detail: 'Request failed' }));
		throw new Error(err.detail || `HTTP ${res.status}`);
	}
	return res.json();
}

export interface DoseScheduleDTO {
	id: string;
	profile_id: string;
	name: string;
	time: string;
	created_at: string;
	updated_at: string;
}

export interface ProfileDTO {
	id: string;
	name: string;
	date_of_birth: string | null;
	timezone: string;
	created_at: string;
	updated_at: string;
	schedules: DoseScheduleDTO[];
}

export interface CreateProfileInput {
	name: string;
	date_of_birth?: string | null;
	timezone: string;
	dose_schedules?: { name: string; time: string }[];
}

export interface UpdateProfileInput {
	name?: string | null;
	date_of_birth?: string | null;
	timezone?: string | null;
}

export interface CreateScheduleInput {
	name: string;
	time: string;
}

export interface UpdateScheduleInput {
	name?: string | null;
	time?: string | null;
}

export async function listProfiles(): Promise<ProfileDTO[]> {
	const res = await fetch('/api/profiles', { headers: authHeaders() });
	const data = await handleResponse<{ profiles: ProfileDTO[] }>(res);
	return data.profiles;
}

export async function getProfile(id: string): Promise<ProfileDTO> {
	const res = await fetch(`/api/profiles/${id}`, { headers: authHeaders() });
	const data = await handleResponse<{ profile: ProfileDTO }>(res);
	return data.profile;
}

export async function createProfile(input: CreateProfileInput): Promise<ProfileDTO> {
	const res = await fetch('/api/profiles', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json', ...authHeaders() },
		body: JSON.stringify(input),
	});
	const data = await handleResponse<{ profile: ProfileDTO }>(res);
	return data.profile;
}

export async function updateProfile(id: string, input: UpdateProfileInput): Promise<ProfileDTO> {
	const res = await fetch(`/api/profiles/${id}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json', ...authHeaders() },
		body: JSON.stringify(input),
	});
	const data = await handleResponse<{ profile: ProfileDTO }>(res);
	return data.profile;
}

export async function deleteProfile(id: string): Promise<void> {
	const res = await fetch(`/api/profiles/${id}`, {
		method: 'DELETE',
		headers: authHeaders(),
	});
	if (!res.ok) {
		const err = await res.json().catch(() => ({ detail: 'Delete failed' }));
		throw new Error(err.detail || `HTTP ${res.status}`);
	}
}

export async function listSchedules(profileId: string): Promise<DoseScheduleDTO[]> {
	const res = await fetch(`/api/profiles/${profileId}/schedules`, { headers: authHeaders() });
	const data = await handleResponse<{ schedules: DoseScheduleDTO[] }>(res);
	return data.schedules;
}

export async function getSchedule(profileId: string, scheduleId: string): Promise<DoseScheduleDTO> {
	const res = await fetch(`/api/profiles/${profileId}/schedules/${scheduleId}`, { headers: authHeaders() });
	const data = await handleResponse<{ schedule: DoseScheduleDTO }>(res);
	return data.schedule;
}

export async function createSchedule(profileId: string, input: CreateScheduleInput): Promise<DoseScheduleDTO> {
	const res = await fetch(`/api/profiles/${profileId}/schedules`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json', ...authHeaders() },
		body: JSON.stringify(input),
	});
	const data = await handleResponse<{ schedule: DoseScheduleDTO }>(res);
	return data.schedule;
}

export async function updateSchedule(
	profileId: string,
	scheduleId: string,
	input: UpdateScheduleInput
): Promise<DoseScheduleDTO> {
	const res = await fetch(`/api/profiles/${profileId}/schedules/${scheduleId}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json', ...authHeaders() },
		body: JSON.stringify(input),
	});
	const data = await handleResponse<{ schedule: DoseScheduleDTO }>(res);
	return data.schedule;
}

export async function deleteSchedule(profileId: string, scheduleId: string): Promise<void> {
	const res = await fetch(`/api/profiles/${profileId}/schedules/${scheduleId}`, {
		method: 'DELETE',
		headers: authHeaders(),
	});
	if (!res.ok) {
		const err = await res.json().catch(() => ({ detail: 'Delete failed' }));
		throw new Error(err.detail || `HTTP ${res.status}`);
	}
}