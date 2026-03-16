<script lang="ts">
	import { api } from '$lib/api';
	import type { Photo } from '$lib/types';

	let photoCount = $state(0);
	let collectionCount = $state(0);
	let recentPhotos = $state<Photo[]>([]);
	let loading = $state(true);

	$effect(() => {
		loadDashboard();
	});

	async function loadDashboard() {
		try {
			const [photosRes, collections] = await Promise.all([
				api.getPhotos(undefined, 6),
				api.getCollections()
			]);
			recentPhotos = photosRes.data;
			photoCount = photosRes.data.length;
			collectionCount = collections.length;
		} catch (e) {
			console.error('Failed to load dashboard:', e);
		} finally {
			loading = false;
		}
	}
</script>

<h1 class="text-2xl font-medium mb-8">Dashboard</h1>

{#if loading}
	<div class="text-text-muted">Loading...</div>
{:else}
	<div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
		<div class="bg-surface border border-border rounded-lg p-6">
			<div class="text-3xl font-mono font-medium">{photoCount}</div>
			<div class="text-sm text-text-muted mt-1">Photos</div>
		</div>
		<div class="bg-surface border border-border rounded-lg p-6">
			<div class="text-3xl font-mono font-medium">{collectionCount}</div>
			<div class="text-sm text-text-muted mt-1">Collections</div>
		</div>
		<div class="bg-surface border border-border rounded-lg p-6">
			<div class="text-3xl font-mono font-medium">
				{(recentPhotos.reduce((sum, p) => sum + p.file_size, 0) / 1024 / 1024).toFixed(1)} MB
			</div>
			<div class="text-sm text-text-muted mt-1">Storage (visible)</div>
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
	{/if}
{/if}
