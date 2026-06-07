<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Alert, AlertDescription, AlertTitle } from '$lib/components/ui/alert';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Dialog } from '$lib/components/ui/dialog';
	import UserRound from '@lucide/svelte/icons/user-round';
	import Globe from '@lucide/svelte/icons/globe';
	import Unlink from '@lucide/svelte/icons/unlink';
	import Link from '@lucide/svelte/icons/link';
	import TriangleAlert from '@lucide/svelte/icons/triangle-alert';
	import Eye from '@lucide/svelte/icons/eye';
	import EyeOff from '@lucide/svelte/icons/eye-off';
	import Mail from '@lucide/svelte/icons/mail';

	const urlParams = $derived($page.url.searchParams);
	const linkedProvider = $derived(urlParams.get('linked'));
	const oauthError = $derived(urlParams.get('oauth_error'));

	let accounts = $state<string[]>([]);
	let hasPassword = $state(false);
	let isLoading = $state(true);
	let unlinkProvider = $state<string | null>(null);
	let isUnlinking = $state(false);
	let unlinkError = $state('');

	let showSetPassword = $state(false);
	let showChangePassword = $state(false);
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let showPassword = $state(false);
	let isSettingPassword = $state(false);
	let passwordError = $state('');

	let showChangeEmail = $state(false);
	let newEmail = $state('');
	let emailPassword = $state('');
	let isChangingEmail = $state(false);
	let emailError = $state('');
	let emailChangeSuccess = $state(false);

	function getToken() {
		return localStorage.getItem('access_token') || '';
	}

	function getCurrentEmail() {
		const token = getToken();
		if (!token) return '';
		try {
			const payload = JSON.parse(atob(token.split('.')[1]));
			return payload.email || '';
		} catch {
			return '';
		}
	}

	async function fetchAccounts() {
		isLoading = true;
		try {
			const res = await fetch('/api/auth/oauth/accounts', {
				headers: { Authorization: `Bearer ${getToken()}` },
			});
			if (res.ok) {
				const data = await res.json();
				accounts = data.accounts?.map((a: { provider: string }) => a.provider) || [];
				hasPassword = data.has_password || false;
			}
		} catch (e) {
			console.error('Failed to fetch accounts', e);
		} finally {
			isLoading = false;
		}
	}

	async function fetchPendingEmailChange() {
		try {
			const res = await fetch('/api/auth/email/change/pending', {
				headers: { Authorization: `Bearer ${getToken()}` },
			});
			if (res.ok) {
				emailChangeSuccess = true;
			}
		} catch (e) {
			console.error('Failed to fetch pending email change', e);
		}
	}

	async function handleLinkOAuth(provider: string) {
		try {
			const res = await fetch(`/api/auth/oauth/${provider}/init`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${getToken()}`,
				},
				body: JSON.stringify({ redirect: '/profile' }),
			});
			if (!res.ok) {
				toast.error('Failed to initiate link');
				return;
			}
			const data = await res.json();
			const state = data.state;
			window.location.href = `/api/auth/oauth/${provider}?state=${encodeURIComponent(state)}`;
		} catch {
			toast.error('Network error');
		}
	}

	async function handleUnlink() {
		if (!unlinkProvider) return;
		isUnlinking = true;
		unlinkError = '';
		try {
			const res = await fetch(`/api/auth/oauth/accounts/${unlinkProvider}`, {
				method: 'DELETE',
				headers: { Authorization: `Bearer ${getToken()}` },
			});
			if (res.ok) {
				toast.success(`${unlinkProvider.charAt(0).toUpperCase() + unlinkProvider.slice(1)} account unlinked`);
				unlinkProvider = null;
				await fetchAccounts();
			} else if (res.status === 403) {
				unlinkError = 'Cannot unlink your last login method. Set a password first.';
			} else {
				const data = await res.json();
				unlinkError = data.detail || 'Failed to unlink';
			}
		} catch {
			unlinkError = 'Network error';
		} finally {
			isUnlinking = false;
		}
	}

	async function handleSetPassword() {
		passwordError = '';
		if (newPassword.length < 8) {
			passwordError = 'Password must be at least 8 characters';
			return;
		}
		if (!/[A-Z]/.test(newPassword) || !/[a-z]/.test(newPassword) || !/[0-9]/.test(newPassword)) {
			passwordError = 'Password must contain uppercase, lowercase, and number';
			return;
		}
		if (newPassword !== confirmPassword) {
			passwordError = 'Passwords do not match';
			return;
		}

		isSettingPassword = true;
		try {
			const res = await fetch('/api/auth/password/set', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${getToken()}`,
				},
				body: JSON.stringify({ password: newPassword }),
			});
			if (res.ok) {
				toast.success('Password set successfully');
				hasPassword = true;
				showSetPassword = false;
				newPassword = '';
				confirmPassword = '';
			} else {
				const data = await res.json();
				passwordError = data.detail || 'Failed to set password';
			}
		} catch {
			passwordError = 'Network error';
		} finally {
			isSettingPassword = false;
		}
	}

	async function handleChangePassword() {
		passwordError = '';
		if (!currentPassword) {
			passwordError = 'Current password is required';
			return;
		}
		if (newPassword.length < 8) {
			passwordError = 'Password must be at least 8 characters';
			return;
		}
		if (!/[A-Z]/.test(newPassword) || !/[a-z]/.test(newPassword) || !/[0-9]/.test(newPassword)) {
			passwordError = 'Password must contain uppercase, lowercase, and number';
			return;
		}
		if (newPassword !== confirmPassword) {
			passwordError = 'Passwords do not match';
			return;
		}
		if (currentPassword === newPassword) {
			passwordError = 'New password must differ from current password';
			return;
		}

		isSettingPassword = true;
		try {
			const res = await fetch('/api/auth/password', {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${getToken()}`,
				},
				body: JSON.stringify({
					current_password: currentPassword,
					new_password: newPassword,
					confirm_password: confirmPassword,
				}),
			});
			if (res.ok) {
				toast.success('Password changed successfully');
				showChangePassword = false;
				currentPassword = '';
				newPassword = '';
				confirmPassword = '';
			} else if (res.status === 403) {
				const data = await res.json();
				passwordError = data.detail || 'Current password is incorrect';
			} else {
				const data = await res.json();
				passwordError = data.detail || 'Failed to change password';
			}
		} catch {
			passwordError = 'Network error';
		} finally {
			isSettingPassword = false;
		}
	}

	async function handleRequestEmailChange() {
		emailError = '';
		if (!newEmail.trim()) {
			emailError = 'New email is required';
			return;
		}
		if (!emailPassword) {
			emailError = 'Current password is required';
			return;
		}
		if (newEmail.trim().toLowerCase() === getCurrentEmail().toLowerCase()) {
			emailError = 'New email must be different from current email';
			return;
		}

		isChangingEmail = true;
		try {
			const res = await fetch('/api/auth/email/change/request', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${getToken()}`,
				},
				body: JSON.stringify({
					new_email: newEmail.trim(),
					current_password: emailPassword,
				}),
			});
			if (res.ok) {
				toast.success('Verification email sent to your new email address');
				emailChangeSuccess = true;
				showChangeEmail = false;
				newEmail = '';
				emailPassword = '';
			} else if (res.status === 409) {
				emailError = 'This email is already in use';
			} else if (res.status === 400) {
				const data = await res.json();
				if (data.detail?.includes('password')) {
					emailError = 'Set a password first before changing your email';
				} else if (data.detail?.includes('incorrect')) {
					emailError = 'Current password is incorrect';
				} else {
					emailError = data.detail || 'Failed to request email change';
				}
			} else {
				const data = await res.json();
				emailError = data.detail || 'Failed to request email change';
			}
		} catch {
			emailError = 'Network error';
		} finally {
			isChangingEmail = false;
		}
	}

	async function handleCancelEmailChange() {
		try {
			const res = await fetch('/api/auth/email/change/cancel', {
				method: 'POST',
				headers: { Authorization: `Bearer ${getToken()}` },
			});
			if (res.ok) {
				toast.success('Email change request cancelled');
				emailChangeSuccess = false;
			} else {
				toast.error('Failed to cancel email change');
			}
		} catch {
			toast.error('Network error');
		}
	}

	function handleLogout() {
		const token = getToken();
		if (token) {
			fetch('/api/auth/logout', {
				method: 'POST',
				headers: { Authorization: `Bearer ${token}` },
			});
		}
		localStorage.removeItem('access_token');
		localStorage.removeItem('refresh_token');
		localStorage.removeItem('email_verified');
		localStorage.removeItem('remember_me');
		goto('/login');
	}

	// Fetch accounts on load
	$effect(() => {
		fetchAccounts();
		fetchPendingEmailChange();
	});

	// Show toast for successful link
	$effect(() => {
		if (linkedProvider) {
			toast.success(`${linkedProvider.charAt(0).toUpperCase() + linkedProvider.slice(1)} account linked successfully`);
			// Clean URL
			const url = new URL(window.location.href);
			url.searchParams.delete('linked');
			window.history.replaceState({}, '', url.toString());
		}
	});

	// Show error toast for OAuth errors
	$effect(() => {
		if (oauthError) {
			const provider = urlParams.get('provider') || '';
			if (oauthError === 'cancelled') {
				toast.error('Linking cancelled');
			} else if (oauthError === 'invalid_state') {
				toast.error('Session expired. Please try again.');
			} else if (oauthError === 'link_failed') {
				toast.error(urlParams.get('oauth_message') || 'Failed to link account');
			} else if (oauthError === 'account_locked') {
				toast.error('Cannot link — set a password first to avoid losing access');
			}
			// Clean URL
			const url = new URL(window.location.href);
			url.searchParams.delete('oauth_error');
			url.searchParams.delete('provider');
			url.searchParams.delete('oauth_message');
			window.history.replaceState({}, '', url.toString());
		}
	});

	const linkedAccounts = $derived(accounts);
	const isGoogleLinked = $derived(linkedAccounts.includes('google'));
</script>

<div class="px-4 py-6">
	<header class="mb-6">
		<h1 class="text-2xl font-semibold tracking-tight">Settings</h1>
		<p class="text-sm text-muted-foreground">Account &amp; security</p>
	</header>

	<!-- OAuth Accounts -->
	<Card class="mb-4">
		<CardHeader>
			<CardTitle class="text-base">Connected Accounts</CardTitle>
			<CardDescription>Link your Google account for quick sign-in</CardDescription>
		</CardHeader>
		<CardContent class="space-y-3">
			<div class="flex items-center justify-between rounded-lg border p-3">
				<div class="flex items-center gap-3">
					<Globe class="size-5 text-muted-foreground" />
					<div>
						<p class="text-sm font-medium">Google</p>
						<p class="text-xs text-muted-foreground">
							{isGoogleLinked ? 'Connected' : 'Not connected'}
						</p>
					</div>
				</div>
				<div class="flex gap-2">
					{#if isGoogleLinked}
						<Button
							variant="outline"
							size="sm"
							onclick={() => (unlinkProvider = 'google')}
						>
							<Unlink class="mr-1 size-3" />
							Unlink
						</Button>
					{:else}
						<Button
							variant="outline"
							size="sm"
							onclick={() => handleLinkOAuth('google')}
							disabled={isLoading}
						>
							<Link class="mr-1 size-3" />
							Link
						</Button>
					{/if}
				</div>
			</div>

			{#if isLoading}
				<p class="text-center text-xs text-muted-foreground">Loading...</p>
			{/if}
		</CardContent>
	</Card>

	<!-- Email Section -->
	<Card class="mb-4">
		<CardHeader>
			<CardTitle class="text-base">Email Address</CardTitle>
			<CardDescription>Update your email address</CardDescription>
		</CardHeader>
		<CardContent class="space-y-3">
			<div class="flex items-center justify-between rounded-lg border p-3">
				<div class="flex items-center gap-3">
					<Mail class="size-5 text-muted-foreground" />
					<div>
						<p class="text-sm font-medium">{getCurrentEmail() || 'Loading...'}</p>
					</div>
				</div>
				{#if emailChangeSuccess}
					<Button variant="outline" size="sm" onclick={handleCancelEmailChange}>
						Cancel Request
					</Button>
				{:else if !showChangeEmail}
					<Button variant="outline" onclick={() => (showChangeEmail = true)}>
						Change Email
					</Button>
				{/if}
			</div>

			{#if emailChangeSuccess}
				<Alert>
					<AlertDescription class="text-sm">
						Email change pending. Check your new inbox for a verification link.
					</AlertDescription>
				</Alert>
			{:else if showChangeEmail}
				<form onsubmit={(e) => { e.preventDefault(); handleRequestEmailChange(); }} class="space-y-3">
					<div class="space-y-1">
						<Label for="new-email">New Email Address</Label>
						<Input
							id="new-email"
							type="email"
							bind:value={newEmail}
							placeholder="your@newemail.com"
						/>
					</div>
					<div class="space-y-1">
						<Label for="email-current-password">Current Password</Label>
						<Input
							id="email-current-password"
							type="password"
							bind:value={emailPassword}
							placeholder="Enter your current password"
						/>
					</div>
					{#if emailError}
						<p class="text-sm text-destructive">{emailError}</p>
					{/if}
					<div class="flex gap-2">
						<Button type="submit" disabled={isChangingEmail}>
							{isChangingEmail ? 'Sending...' : 'Send Verification'}
						</Button>
						<Button
							type="button"
							variant="ghost"
							onclick={() => { showChangeEmail = false; emailError = ''; newEmail = ''; emailPassword = ''; }}
						>
							Cancel
						</Button>
					</div>
				</form>
			{/if}
		</CardContent>
	</Card>

	<!-- Password Section -->
	<Card class="mb-4">
		<CardHeader>
			<CardTitle class="text-base">Password</CardTitle>
			<CardDescription>
				{hasPassword ? 'Password is set' : 'Set a password for email sign-in'}
			</CardDescription>
		</CardHeader>
		<CardContent>
			{#if hasPassword}
				{#if showChangePassword}
					<form onsubmit={(e) => { e.preventDefault(); handleChangePassword(); }} class="space-y-3">
						<div class="space-y-1">
							<Label for="current-password">Current Password</Label>
							<Input
								id="current-password"
								type="password"
								bind:value={currentPassword}
								placeholder="Enter your current password"
							/>
						</div>
						<div class="space-y-1">
							<Label for="change-new-password">New Password</Label>
							<div class="relative">
								<Input
									id="change-new-password"
									type={showPassword ? 'text' : 'password'}
									bind:value={newPassword}
									class="pr-10"
									placeholder="8+ chars, upper+lower+number"
								/>
								<button
									type="button"
									onclick={() => (showPassword = !showPassword)}
									class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
								>
									{#if showPassword}
										<EyeOff class="h-4 w-4" />
									{:else}
										<Eye class="h-4 w-4" />
									{/if}
								</button>
							</div>
						</div>
						<div class="space-y-1">
							<Label for="change-confirm-password">Confirm Password</Label>
							<Input
								id="change-confirm-password"
								type="password"
								bind:value={confirmPassword}
								placeholder="Repeat your password"
							/>
						</div>
						{#if passwordError}
							<p class="text-sm text-destructive">{passwordError}</p>
						{/if}
						<div class="flex gap-2">
							<Button type="submit" disabled={isSettingPassword}>
								{isSettingPassword ? 'Changing...' : 'Change Password'}
							</Button>
							<Button
								type="button"
								variant="ghost"
								onclick={() => { showChangePassword = false; passwordError = ''; currentPassword = ''; newPassword = ''; confirmPassword = ''; }}
							>
								Cancel
							</Button>
						</div>
					</form>
				{:else}
					<div class="flex items-center justify-between">
						<p class="text-sm text-muted-foreground">Your account has a password.</p>
						<Button variant="outline" onclick={() => (showChangePassword = true)}>
							Change Password
						</Button>
					</div>
				{/if}
			{:else if showSetPassword}
				<form onsubmit={(e) => { e.preventDefault(); handleSetPassword(); }} class="space-y-3">
					<div class="space-y-1">
						<Label for="new-password">New Password</Label>
						<div class="relative">
							<Input
								id="new-password"
								type={showPassword ? 'text' : 'password'}
								bind:value={newPassword}
								class="pr-10"
								placeholder="8+ chars, upper+lower+number"
							/>
							<button
								type="button"
								onclick={() => (showPassword = !showPassword)}
								class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
							>
								{#if showPassword}
									<EyeOff class="h-4 w-4" />
								{:else}
									<Eye class="h-4 w-4" />
								{/if}
							</button>
						</div>
					</div>
					<div class="space-y-1">
						<Label for="confirm-password">Confirm Password</Label>
						<Input
							id="confirm-password"
							type="password"
							bind:value={confirmPassword}
							placeholder="Repeat your password"
						/>
					</div>
					{#if passwordError}
						<p class="text-sm text-destructive">{passwordError}</p>
					{/if}
					<div class="flex gap-2">
						<Button type="submit" disabled={isSettingPassword}>
							{isSettingPassword ? 'Setting...' : 'Set Password'}
						</Button>
						<Button
							type="button"
							variant="ghost"
							onclick={() => { showSetPassword = false; passwordError = ''; }}
						>
							Cancel
						</Button>
					</div>
				</form>
			{:else}
				<Button variant="outline" onclick={() => (showSetPassword = true)}>
					Set Password
				</Button>
			{/if}
		</CardContent>
	</Card>

	<!-- Sign Out -->
	<Button variant="ghost" class="w-full text-destructive" onclick={handleLogout}>
		Sign Out
	</Button>
</div>

<!-- Unlink Confirmation Dialog -->
<Dialog
	open={unlinkProvider !== null}
	onclose={() => { unlinkProvider = null; unlinkError = ''; }}
>
	<div class="space-y-4">
		<div class="space-y-2">
			<h2 class="text-lg font-semibold">Unlink {unlinkProvider} Account</h2>
			<p class="text-sm text-muted-foreground">
				Are you sure you want to unlink your {unlinkProvider} account?
			</p>
		</div>

		{#if !hasPassword && accounts.length <= 1}
			<p class="text-sm text-destructive">
				You don't have a password set. Add a password before unlinking your only login method.
			</p>
		{:else if !hasPassword}
			<Alert variant="destructive">
				<TriangleAlert class="h-4 w-4" />
				<AlertTitle>Warning</AlertTitle>
				<AlertDescription>
					You don't have a password set. If you unlink this account, you'll have one login method remaining.
				</AlertDescription>
			</Alert>
		{/if}

		{#if unlinkError}
			<p class="text-sm text-destructive">{unlinkError}</p>
		{/if}

		<div class="flex justify-end gap-2">
			<Button
				variant="outline"
				onclick={() => { unlinkProvider = null; unlinkError = ''; }}
			>
				{!hasPassword && accounts.length <= 1 ? 'Close' : 'Cancel'}
			</Button>
			{#if !(!hasPassword && accounts.length <= 1)}
				<Button
					variant="destructive"
					onclick={handleUnlink}
					disabled={isUnlinking}
				>
					{isUnlinking ? 'Unlinking...' : 'Unlink'}
				</Button>
			{/if}
		</div>
	</div>
</Dialog>
