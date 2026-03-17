<script lang="ts">
	import { api } from '$lib/api';
	import type { Photo } from '$lib/types';
	import type { AdminStats } from '$lib/types';

	let stats = $state<AdminStats | null>(null);
	let recentPhotos = $state<Photo[]>([]);
	let loading = $state(true);

	$effect(() => {
		loadDashboard();
	});

	async function loadDashboard() {
		try {
			const [adminStats, photosRes] = await Promise.all([
				api.getAdminStats(),
				api.getPhotos(undefined, 6)
			]);
			stats = adminStats;
			recentPhotos = photosRes.data;
		} catch (e) {
			console.error('Failed to load dashboard:', e);
		} finally {
			loading = false;
		}
	}

	function formatBytes(bytes: number): string {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
		return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GB`;
	}
</script>

<h1 class="text-2xl font-medium mb-8">Dashboard</h1>

{#if loading}
	<div class="text-text-muted">Loading...</div>
{:else if stats}
	<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
		<div class="bg-surface border border-border rounded-lg p-6">
			<div class="text-3xl font-mono font-medium">{stats.roll_count}</div>
			<div class="text-sm text-text-muted mt-1">Rolls</div>
		</div>
		<div class="bg-surface border border-border rounded-lg p-6">
			<div class="text-3xl font-mono font-medium">{stats.photo_count}</div>
			<div class="text-sm text-text-muted mt-1">Photos</div>
		</div>
		<div class="bg-surface border border-border rounded-lg p-6">
			<div class="text-3xl font-mono font-medium">{stats.collection_count}</div>
			<div class="text-sm text-text-muted mt-1">Collections</div>
		</div>
		<div class="bg-surface border border-border rounded-lg p-6">
			<div class="text-3xl font-mono font-medium">{formatBytes(stats.storage_bytes)}</div>
			<div class="text-sm text-text-muted mt-1">Storage</div>
		</div>
	</div>

	{#if recentPhotos.length > 0}
		<h2 class="text-lg font-medium mb-4">Recent Photos</h2>
		<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-2">
			{#each recentPhotos as photo}
				<div class="aspect-square bg-surface rounded-md overflow-hidden">
					<img
						src={photo.urls.thumb}
						alt={photo.title || ''}
						class="w-full h-full object-cover"
					/>
				</div>
			{/each}
		</div>
	{:else if stats.roll_count === 0}
		<div class="text-center py-16 bg-surface border border-border rounded-lg">
			<p class="text-text-muted mb-4">No rolls yet. Start by creating your first roll.</p>
			<a
				href="/admin/rolls"
				class="px-4 py-2 bg-amber-600 hover:bg-amber-500 text-white rounded-md text-sm font-medium transition-colors inline-block"
			>
				Go to Rolls
			</a>
		</div>
	{/if}
{/if}
