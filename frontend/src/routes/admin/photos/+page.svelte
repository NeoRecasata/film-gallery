<script lang="ts">
	import { api } from '$lib/api';
	import type { Photo } from '$lib/types';
	import UploadZone from '$lib/components/admin/UploadZone.svelte';
	import PhotoEditor from '$lib/components/admin/PhotoEditor.svelte';

	let photos = $state<Photo[]>([]);
	let loading = $state(true);
	let filter = $state<'all' | 'published' | 'draft'>('all');

	const filtered = $derived.by(() => {
		if (filter === 'published') return photos.filter(p => p.published);
		if (filter === 'draft') return photos.filter(p => !p.published);
		return photos;
	});

	$effect(() => {
		loadPhotos();
	});

	async function loadPhotos() {
		loading = true;
		try {
			const res = await api.getPhotos(undefined, 100);
			photos = res.data;
		} catch (e) {
			console.error('Failed to load photos:', e);
		} finally {
			loading = false;
		}
	}

	function handleUploaded(photo: Photo) {
		photos = [photo, ...photos];
	}

	function handleUpdate(updated: Photo) {
		photos = photos.map(p => p.id === updated.id ? updated : p);
	}

	function handleDelete(id: string) {
		photos = photos.filter(p => p.id !== id);
	}
</script>

<div class="max-w-4xl">
	<h1 class="text-2xl font-medium mb-6">Photos</h1>

	<UploadZone onuploaded={handleUploaded} />

	<div class="mt-8">
		<div class="flex items-center gap-4 mb-4">
			<h2 class="text-lg font-medium">Library</h2>
			<div class="flex gap-1 text-sm">
				{#each ['all', 'published', 'draft'] as f}
					<button
						onclick={() => filter = f as typeof filter}
						class="px-3 py-1 rounded transition-colors
							{filter === f ? 'bg-surface-hover text-text' : 'text-text-muted hover:text-text'}"
					>
						{f.charAt(0).toUpperCase() + f.slice(1)}
					</button>
				{/each}
			</div>
		</div>

		{#if loading}
			<p class="text-text-muted">Loading photos...</p>
		{:else if filtered.length === 0}
			<p class="text-text-muted py-8 text-center">No photos found.</p>
		{:else}
			<div class="space-y-2">
				{#each filtered as photo (photo.id)}
					<PhotoEditor
						{photo}
						onupdate={handleUpdate}
						ondelete={handleDelete}
					/>
				{/each}
			</div>
		{/if}
	</div>
</div>
