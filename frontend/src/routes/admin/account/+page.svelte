<script lang="ts">
	import { api } from '$lib/api';

	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let error = $state('');
	let success = $state(false);
	let loading = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		success = false;

		if (newPassword !== confirmPassword) {
			error = 'New passwords do not match';
			return;
		}
		if (newPassword.length < 8) {
			error = 'Password must be at least 8 characters';
			return;
		}

		loading = true;
		try {
			await api.changePassword(currentPassword, newPassword);
			success = true;
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to change password';
		} finally {
			loading = false;
		}
	}
</script>

<div class="max-w-md">
	<h1 class="text-2xl font-medium mb-6">Account</h1>

	<div class="bg-surface border border-border rounded-lg p-6">
		<h2 class="font-medium mb-4">Change Password</h2>

		{#if error}
			<div class="mb-4 p-3 bg-error/10 border border-error/20 rounded-md text-error text-sm">{error}</div>
		{/if}
		{#if success}
			<div class="mb-4 p-3 bg-success/10 border border-success/20 rounded-md text-success text-sm">
				Password changed successfully. You will need to log in again on other devices.
			</div>
		{/if}

		<form onsubmit={handleSubmit} class="space-y-4">
			<div>
				<label for="current" class="block text-sm text-text-muted mb-1">Current Password</label>
				<input id="current" type="password" bind:value={currentPassword} required class="w-full px-3 py-2 bg-bg border border-border rounded-md focus:outline-none focus:border-accent" />
			</div>
			<div>
				<label for="new" class="block text-sm text-text-muted mb-1">New Password</label>
				<input id="new" type="password" bind:value={newPassword} required minlength="8" class="w-full px-3 py-2 bg-bg border border-border rounded-md focus:outline-none focus:border-accent" />
			</div>
			<div>
				<label for="confirm" class="block text-sm text-text-muted mb-1">Confirm New Password</label>
				<input id="confirm" type="password" bind:value={confirmPassword} required class="w-full px-3 py-2 bg-bg border border-border rounded-md focus:outline-none focus:border-accent" />
			</div>
			<button type="submit" disabled={loading} class="px-4 py-2 bg-text text-bg rounded-md text-sm font-medium disabled:opacity-50">
				{loading ? 'Changing...' : 'Change Password'}
			</button>
		</form>
	</div>
</div>
