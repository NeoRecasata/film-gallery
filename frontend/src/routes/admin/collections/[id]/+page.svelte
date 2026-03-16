<script lang="ts">
	import { page } from '$app/state';
	import { api } from '$lib/api';
	import type { Collection, Photo } from '$lib/types';

	let collectionId = $derived(page.params.id);
	let collection = $state<Collection | null>(null);
	let allPhotos = $state<Photo[]>([]);
	let loading = $state(true);
	let saving = $state(false);

	let title = $state('');
	let description = $state('');

	const collectionPhotoIds = $derived.by(() => {
		return new Set(collection?.photos?.map(p => p.id) || []);
	});
	const availablePhotos = $derived.by(() => {
		return allPhotos.filter(p => !collectionPhotoIds.has(p.id));
	});

	$effect(() => {
		loadData();
	});

	async function loadData() {
		try {
			const [colls, photosRes] = await Promise.all([
				api.getCollections(),
				api.getPhotos(undefined, 100)
			]);
			// Find collection by ID, then load its detail by slug
			const found = colls.find(c => c.id === collectionId);
			if (found) {
				const detail = await api.getCollection(found.slug);
				collection = detail;
				title = detail.title;
				description = detail.description || '';
			}
			allPhotos = photosRes.data;
		} catch (e) {
			console.error('Failed to load:', e);
		} finally {
			loading = false;
		}
	}

	async function addPhoto(photoId: string) {
		if (!collection) return;
		const currentPhotos = collection.photos || [];
		const newPhotos = [
			...currentPhotos.map((p, i) => ({ photo_id: p.id, sort_order: i })),
			{ photo_id: photoId, sort_order: currentPhotos.length }
		];
		try {
			await api.setCollectionPhotos(collection.id, newPhotos);
			await loadData(); // reload to get updated photos
		} catch (e) {
			console.error('Failed to add photo:', e);
		}
	}

	async function removePhoto(photoId: string) {
		if (!collection) return;
		const remaining = (collection.photos || [])
			.filter(p => p.id !== photoId)
			.map((p, i) => ({ photo_id: p.id, sort_order: i }));
		try {
			await api.setCollectionPhotos(collection.id, remaining);
			await loadData();
		} catch (e) {
			console.error('Failed to remove photo:', e);
		}
	}

	async function saveDetails() {
		if (!collection) return;
		saving = true;
		try {
			const updated = await api.updateCollection(collection.id, {
				title: title.trim(),
				description: description.trim() || null
			} as Partial<Collection>);
			collection = updated;
		} catch (e) {
			console.error('Failed to update:', e);
		} finally {
			saving = false;
		}
	}
</script>

<div class="max-w-4xl">
	{#if loading}
		<p class="text-text-muted">Loading...</p>
	{:else if !collection}
		<p class="text-error">Collection not found.</p>
	{:else}
		<div class="flex items-center gap-4 mb-6">
			<a href="/admin/collections" class="text-text-muted hover:text-text text-sm">&larr; Back</a>
			<h1 class="text-2xl font-medium">Edit: {collection.title}</h1>
		</div>

		<div class="bg-surface border border-border rounded-lg p-6 space-y-4">
			<div>
				<label for="title" class="block text-sm text-text-muted mb-1">Title</label>
				<input
					id="title"
					bind:value={title}
					class="w-full px-3 py-2 bg-bg border border-border rounded-md focus:outline-none focus:border-accent"
				/>
			</div>
			<div>
				<label for="desc" class="block text-sm text-text-muted mb-1">Description</label>
				<textarea
					id="desc"
					bind:value={description}
					rows="3"
					class="w-full px-3 py-2 bg-bg border border-border rounded-md focus:outline-none focus:border-accent resize-none"
				></textarea>
			</div>
			<button
				onclick={saveDetails}
				disabled={saving}
				class="px-4 py-2 bg-text text-bg rounded-md text-sm font-medium disabled:opacity-50"
			>
				{saving ? 'Saving...' : 'Save'}
			</button>
		</div>

		<div class="mt-8">
			<h2 class="text-lg font-medium mb-4">Photos in this collection</h2>

			{#if collection.photos && collection.photos.length > 0}
				<div class="grid grid-cols-4 sm:grid-cols-6 gap-2">
					{#each collection.photos as photo}
						<div class="aspect-square bg-surface rounded overflow-hidden relative group">
							<img src={photo.urls.thumb} alt={photo.title || ''} class="w-full h-full object-cover" />
							<button
								onclick={() => removePhoto(photo.id)}
								class="absolute top-1 right-1 bg-black/60 text-white rounded-full w-5 h-5 text-xs opacity-0 group-hover:opacity-100 transition-opacity"
							>&times;</button>
						</div>
					{/each}
				</div>
			{:else}
				<p class="text-text-muted text-sm">No photos added to this collection yet.</p>
			{/if}

			<h3 class="text-sm font-medium mt-6 mb-2 text-text-muted">Add photos</h3>
			<div class="grid grid-cols-4 sm:grid-cols-8 gap-2">
				{#each availablePhotos as photo}
					<button
						onclick={() => addPhoto(photo.id)}
						class="aspect-square bg-surface rounded overflow-hidden opacity-60 hover:opacity-100 transition-opacity"
					>
						<img src={photo.urls.thumb} alt={photo.title || ''} class="w-full h-full object-cover" />
					</button>
				{/each}
			</div>
		</div>
	{/if}
</div>
