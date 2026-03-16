<script lang="ts">
	import { api } from '$lib/api';
	import type { Photo } from '$lib/types';

	let {
		photo,
		onupdate,
		ondelete
	}: {
		photo: Photo;
		onupdate: (photo: Photo) => void;
		ondelete: (id: string) => void;
	} = $props();

	let editing = $state(false);
	let title = $state(photo.title || '');
	let description = $state(photo.description || '');
	let filmStock = $state(photo.film_stock || '');
	let camera = $state(photo.camera || '');
	let lens = $state(photo.lens || '');
	let saving = $state(false);

	async function save() {
		saving = true;
		try {
			const updated = await api.updatePhoto(photo.id, {
				title: title || null,
				description: description || null,
				film_stock: filmStock || null,
				camera: camera || null,
				lens: lens || null
			} as Partial<Photo>);
			onupdate(updated);
			editing = false;
		} catch (e) {
			console.error('Failed to update photo:', e);
		} finally {
			saving = false;
		}
	}

	async function togglePublished() {
		try {
			const updated = await api.updatePhoto(photo.id, { published: !photo.published } as Partial<Photo>);
			onupdate(updated);
		} catch (e) {
			console.error('Failed to toggle published:', e);
		}
	}

	async function handleDelete() {
		if (!confirm('Delete this photo? This cannot be undone.')) return;
		try {
			await api.deletePhoto(photo.id);
			ondelete(photo.id);
		} catch (e) {
			console.error('Failed to delete photo:', e);
		}
	}
</script>

<div class="bg-surface border border-border rounded-lg overflow-hidden">
	<div class="flex">
		<div class="w-32 h-32 flex-shrink-0">
			<img src={photo.urls.thumb} alt={photo.title || ''} class="w-full h-full object-cover" />
		</div>

		<div class="flex-1 p-3 min-w-0">
			{#if editing}
				<div class="space-y-2">
					<input bind:value={title} placeholder="Title" class="w-full px-2 py-1 bg-bg border border-border rounded text-sm" />
					<input bind:value={description} placeholder="Description" class="w-full px-2 py-1 bg-bg border border-border rounded text-sm" />
					<div class="grid grid-cols-3 gap-2">
						<input bind:value={filmStock} placeholder="Film stock" class="px-2 py-1 bg-bg border border-border rounded text-sm" />
						<input bind:value={camera} placeholder="Camera" class="px-2 py-1 bg-bg border border-border rounded text-sm" />
						<input bind:value={lens} placeholder="Lens" class="px-2 py-1 bg-bg border border-border rounded text-sm" />
					</div>
					<div class="flex gap-2">
						<button onclick={save} disabled={saving} class="px-3 py-1 bg-text text-bg rounded text-sm disabled:opacity-50">
							{saving ? 'Saving...' : 'Save'}
						</button>
						<button onclick={() => editing = false} class="px-3 py-1 text-text-muted text-sm">Cancel</button>
					</div>
				</div>
			{:else}
				<div class="flex items-start justify-between">
					<div>
						<p class="font-medium text-sm">{photo.title || photo.slug}</p>
						<p class="text-xs text-text-muted mt-1">
							{photo.width}x{photo.height} &middot; {(photo.file_size / 1024 / 1024).toFixed(1)} MB
						</p>
						{#if photo.film_stock || photo.camera}
							<p class="text-xs text-text-muted">
								{[photo.film_stock, photo.camera, photo.lens].filter(Boolean).join(' / ')}
							</p>
						{/if}
					</div>
					<div class="flex items-center gap-2">
						<button
							onclick={togglePublished}
							class="px-2 py-1 rounded text-xs border transition-colors
								{photo.published ? 'border-success/30 text-success' : 'border-border text-text-muted'}"
						>
							{photo.published ? 'Published' : 'Draft'}
						</button>
						<button onclick={() => editing = true} class="text-xs text-text-muted hover:text-text">Edit</button>
						<button onclick={handleDelete} class="text-xs text-error/60 hover:text-error">Delete</button>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>
