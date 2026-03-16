<script lang="ts">
	import { api, setAccessToken } from '$lib/api';
	import { goto } from '$app/navigation';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;
		try {
			const res = await api.login(email, password);
			setAccessToken(res.access_token);
			goto('/admin');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Login failed';
		} finally {
			loading = false;
		}
	}
</script>

<div class="min-h-screen flex items-center justify-center">
	<div class="w-full max-w-sm p-8">
		<h1 class="text-2xl font-medium mb-6">Admin Login</h1>

		{#if error}
			<div class="mb-4 p-3 bg-error/10 border border-error/20 rounded-md text-error text-sm">
				{error}
			</div>
		{/if}

		<form onsubmit={handleSubmit} class="space-y-4">
			<div>
				<label for="email" class="block text-sm text-text-muted mb-1">Email</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					required
					class="w-full px-3 py-2 bg-surface border border-border rounded-md text-text focus:outline-none focus:border-accent"
				/>
			</div>
			<div>
				<label for="password" class="block text-sm text-text-muted mb-1">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					class="w-full px-3 py-2 bg-surface border border-border rounded-md text-text focus:outline-none focus:border-accent"
				/>
			</div>
			<button
				type="submit"
				disabled={loading}
				class="w-full py-2 bg-text text-bg rounded-md font-medium hover:bg-text/90 transition-colors disabled:opacity-50"
			>
				{loading ? 'Logging in...' : 'Log in'}
			</button>
		</form>
	</div>
</div>
