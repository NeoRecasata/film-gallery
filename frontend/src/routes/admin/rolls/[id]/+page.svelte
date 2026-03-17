<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import type { Roll, Photo } from '$lib/types';
	import UploadZone from '$lib/components/admin/UploadZone.svelte';

	let rollId = $derived(page.params.id as string);
	let roll = $state<Roll | null>(null);
	let loading = $state(true);
	let saving = $state(false);
	let deleting = $state(false);

	// Form state
	let title = $state('');
	let slug = $state('');
	let description = $state('');
	let camera = $state('');
	let filmStock = $state('');
	let lens = $state('');
	let location = $state('');
	let shotAt = $state('');
	let published = $state(false);

	// Photo editor state
	let selectedPhotoId = $state<string | null>(null);
	let photoTitle = $state('');
	let photoDescription = $state('');
	let photoCamera = $state('');
	let photoFilmStock = $state('');
	let photoLens = $state('');
	let photoLocation = $state('');
	let photoTakenAt = $state('');
	let photoHidden = $state(false);
	let savingPhoto = $state(false);
	let deletingPhoto = $state(false);

	const photos = $derived(roll?.photos || []);
	const selectedPhoto = $derived(photos.find(p => p.id === selectedPhotoId) || null);

	$effect(() => {
		loadRoll();
	});

	// Sync photo editor fields when selection changes
	$effect(() => {
		if (selectedPhoto) {
			photoTitle = selectedPhoto.title || '';
			photoDescription = selectedPhoto.description || '';
			photoCamera = selectedPhoto.camera || '';
			photoFilmStock = selectedPhoto.film_stock || '';
			photoLens = selectedPhoto.lens || '';
			photoLocation = selectedPhoto.location || '';
			photoTakenAt = selectedPhoto.taken_at ? selectedPhoto.taken_at.split('T')[0] : '';
			photoHidden = selectedPhoto.hidden;
		}
	});

	async function loadRoll() {
		loading = true;
		try {
			const data = await api.getRoll(rollId);
			roll = data;
			title = data.title;
			slug = data.slug;
			description = data.description || '';
			camera = data.camera || '';
			filmStock = data.film_stock || '';
			lens = data.lens || '';
			location = data.location || '';
			shotAt = data.shot_at ? data.shot_at.split('T')[0] : '';
			published = data.published;
		} catch (e) {
			console.error('Failed to load roll:', e);
		} finally {
			loading = false;
		}
	}

	async function saveRoll() {
		if (!roll) return;
		saving = true;
		try {
			const updated = await api.updateRoll(roll.id, {
				title: title.trim(),
				slug: slug.trim(),
				description: description.trim() || null,
				camera: camera.trim() || null,
				film_stock: filmStock.trim() || null,
				lens: lens.trim() || null,
				location: location.trim() || null,
				shot_at: shotAt || null,
				published
			});
			roll = { ...roll, ...updated };
		} catch (e) {
			console.error('Failed to save roll:', e);
		} finally {
			saving = false;
		}
	}

	async function deleteRoll() {
		if (!roll) return;
		if (!confirm('Delete this roll and all its photos? This cannot be undone.')) return;
		deleting = true;
		try {
			await api.deleteRoll(roll.id);
			goto('/admin/rolls');
		} catch (e) {
			console.error('Failed to delete roll:', e);
			deleting = false;
		}
	}

	function handleUploaded(photo: Photo) {
		if (roll) {
			roll = { ...roll, photos: [...(roll.photos || []), photo], photo_count: (roll.photo_count ?? 0) + 1 };
		}
	}

	function selectPhoto(id: string) {
		selectedPhotoId = selectedPhotoId === id ? null : id;
	}

	async function savePhoto() {
		if (!selectedPhoto) return;
		savingPhoto = true;
		try {
			const updated = await api.updatePhoto(selectedPhoto.id, {
				title: photoTitle.trim() || null,
				description: photoDescription.trim() || null,
				camera: photoCamera.trim() || null,
				film_stock: photoFilmStock.trim() || null,
				lens: photoLens.trim() || null,
				location: photoLocation.trim() || null,
				taken_at: photoTakenAt || null,
				hidden: photoHidden
			} as Partial<Photo>);
			if (roll) {
				roll = { ...roll, photos: roll.photos?.map(p => p.id === updated.id ? updated : p) };
			}
		} catch (e) {
			console.error('Failed to save photo:', e);
		} finally {
			savingPhoto = false;
		}
	}

	async function togglePhotoHidden() {
		if (!selectedPhoto) return;
		photoHidden = !photoHidden;
		try {
			const updated = await api.updatePhoto(selectedPhoto.id, { hidden: photoHidden } as Partial<Photo>);
			if (roll) {
				roll = { ...roll, photos: roll.photos?.map(p => p.id === updated.id ? updated : p) };
			}
		} catch (e) {
			console.error('Failed to toggle hidden:', e);
			photoHidden = !photoHidden;
		}
	}

	async function setAsCover() {
		if (!roll || !selectedPhotoId) return;
		try {
			const updated = await api.updateRoll(roll.id, { cover_photo_id: selectedPhotoId });
			roll = { ...roll, ...updated };
		} catch (e) {
			console.error('Failed to set cover:', e);
		}
	}

	async function deletePhoto() {
		if (!selectedPhoto) return;
		if (!confirm('Delete this photo? This cannot be undone.')) return;
		deletingPhoto = true;
		try {
			await api.deletePhoto(selectedPhoto.id);
			if (roll) {
				roll = {
					...roll,
					photos: roll.photos?.filter(p => p.id !== selectedPhoto.id),
					photo_count: Math.max(0, (roll.photo_count ?? 1) - 1)
				};
			}
			selectedPhotoId = null;
		} catch (e) {
			console.error('Failed to delete photo:', e);
		} finally {
			deletingPhoto = false;
		}
	}
</script>

{#if loading}
	<p class="text-text-muted">Loading...</p>
{:else if !roll}
	<p class="text-error">Roll not found.</p>
{:else}
	<!-- Top bar -->
	<div class="flex items-center justify-between mb-6">
		<div class="flex items-center gap-2 text-sm">
			<a href="/admin/rolls" class="text-text-muted hover:text-text transition-colors">&larr; Rolls</a>
			<span class="text-text-muted/40">/</span>
			<span class="text-text">{roll.title}</span>
		</div>
		<div class="flex items-center gap-3">
			<button
				onclick={deleteRoll}
				disabled={deleting}
				class="px-3 py-1.5 border border-error/30 text-error/70 hover:text-error hover:border-error/60 rounded-md text-sm transition-colors disabled:opacity-50"
			>
				{deleting ? 'Deleting...' : 'Delete Roll'}
			</button>
			<button
				onclick={saveRoll}
				disabled={saving}
				class="px-4 py-1.5 bg-amber-600 hover:bg-amber-500 text-white rounded-md text-sm font-medium transition-colors disabled:opacity-50"
			>
				{saving ? 'Saving...' : 'Save'}
			</button>
		</div>
	</div>

	<div class="flex flex-col lg:flex-row gap-6">
		<!-- Left sidebar - metadata -->
		<div class="w-full lg:w-[280px] flex-shrink-0">
			<div class="bg-surface border border-border rounded-lg p-5 space-y-4">
				<h2 class="text-[11px] font-semibold uppercase tracking-wider text-text-muted">Metadata</h2>

				<div>
					<label for="roll-title" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Title</label>
					<input
						id="roll-title"
						bind:value={title}
						class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent"
					/>
				</div>

				<div>
					<label for="roll-slug" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Slug</label>
					<input
						id="roll-slug"
						bind:value={slug}
						class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent"
					/>
				</div>

				<div>
					<label for="roll-desc" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Description</label>
					<textarea
						id="roll-desc"
						bind:value={description}
						rows="3"
						class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent resize-none"
					></textarea>
				</div>

				<div>
					<label for="roll-camera" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Camera</label>
					<input
						id="roll-camera"
						bind:value={camera}
						placeholder="Camera"
						class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40"
					/>
				</div>

				<div>
					<label for="roll-film" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Film Stock</label>
					<input
						id="roll-film"
						bind:value={filmStock}
						placeholder="Film Stock"
						class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40"
					/>
				</div>

				<div>
					<label for="roll-lens" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Lens</label>
					<input
						id="roll-lens"
						bind:value={lens}
						placeholder="Lens"
						class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40"
					/>
				</div>

				<div>
					<label for="roll-location" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Location</label>
					<input
						id="roll-location"
						bind:value={location}
						placeholder="Location"
						class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40"
					/>
				</div>

				<div>
					<label for="roll-date" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Shot Date</label>
					<input
						id="roll-date"
						type="date"
						bind:value={shotAt}
						class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent"
					/>
				</div>

				<div class="flex items-center justify-between pt-2">
					<label for="roll-published" class="text-[11px] uppercase tracking-wide text-text-muted">Published</label>
					<button
						id="roll-published"
						type="button"
						aria-label="Toggle published"
						onclick={() => published = !published}
						class="relative w-10 h-5 rounded-full transition-colors {published ? 'bg-success' : 'bg-surface-hover'}"
					>
						<span class="absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full transition-transform {published ? 'translate-x-5' : ''}"></span>
					</button>
				</div>
			</div>
		</div>

		<!-- Right area - photos -->
		<div class="flex-1 min-w-0">
			<!-- Upload zone -->
			<UploadZone rollId={roll.id} onuploaded={handleUploaded} />

			<!-- Photo grid -->
			<div class="mt-6">
				<h2 class="text-[11px] font-semibold uppercase tracking-wider text-text-muted mb-3">
					Photos ({photos.length})
				</h2>

				{#if photos.length === 0}
					<p class="text-text-muted text-sm py-8 text-center">No photos yet. Upload some above.</p>
				{:else}
					<div class="grid grid-cols-3 sm:grid-cols-4 lg:grid-cols-5 gap-2">
						{#each photos as photo (photo.id)}
							<button
								onclick={() => selectPhoto(photo.id)}
								class="aspect-square rounded overflow-hidden relative group
									{selectedPhotoId === photo.id ? 'ring-2 ring-amber-500' : ''}"
							>
								<img
									src={photo.urls.thumb}
									alt={photo.title || ''}
									class="w-full h-full object-cover"
								/>
								{#if photo.hidden}
									<span class="absolute top-1 left-1 px-1 py-0.5 bg-error/80 text-white text-[9px] font-semibold uppercase rounded">
										Hidden
									</span>
								{/if}
								{#if photo.id === roll.cover_photo_id}
									<span class="absolute top-1 right-1 px-1 py-0.5 bg-amber-500/80 text-white text-[9px] font-semibold uppercase rounded">
										Cover
									</span>
								{/if}
							</button>
						{/each}
					</div>
				{/if}
			</div>

			<!-- Photo editor -->
			{#if selectedPhoto}
				<div class="mt-6 bg-surface border border-border rounded-lg p-5">
					<div class="flex flex-col sm:flex-row gap-5">
						<!-- Preview -->
						<div class="w-full sm:w-48 flex-shrink-0">
							<img
								src={selectedPhoto.urls.medium || selectedPhoto.urls.thumb}
								alt={selectedPhoto.title || ''}
								class="w-full rounded object-contain"
							/>
						</div>

						<!-- Fields -->
						<div class="flex-1 space-y-3">
							<div>
								<label for="photo-title" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Title</label>
								<input
									id="photo-title"
									bind:value={photoTitle}
									placeholder="Optional title"
									class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40"
								/>
							</div>

							<div>
								<label for="photo-desc" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Description</label>
								<textarea
									id="photo-desc"
									bind:value={photoDescription}
									rows="2"
									class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent resize-none placeholder:text-text-muted/40"
								></textarea>
							</div>

							<div class="grid grid-cols-2 gap-3">
								<div>
									<label for="photo-camera" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Camera</label>
									<input
										id="photo-camera"
										bind:value={photoCamera}
										placeholder={camera || 'Camera'}
										class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40"
									/>
								</div>
								<div>
									<label for="photo-film" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Film Stock</label>
									<input
										id="photo-film"
										bind:value={photoFilmStock}
										placeholder={filmStock || 'Film Stock'}
										class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40"
									/>
								</div>
								<div>
									<label for="photo-lens" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Lens</label>
									<input
										id="photo-lens"
										bind:value={photoLens}
										placeholder={lens || 'Lens'}
										class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40"
									/>
								</div>
								<div>
									<label for="photo-location" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Location</label>
									<input
										id="photo-location"
										bind:value={photoLocation}
										placeholder={location || 'Location'}
										class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40"
									/>
								</div>
							</div>

							<div>
								<label for="photo-date" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Date Taken</label>
								<input
									id="photo-date"
									type="date"
									bind:value={photoTakenAt}
									class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent"
								/>
							</div>

							<div class="flex items-center gap-4 pt-2">
								<button
									onclick={togglePhotoHidden}
									class="px-3 py-1.5 rounded-md text-xs border transition-colors
										{photoHidden ? 'border-error/30 text-error/70' : 'border-success/30 text-success'}"
								>
									{photoHidden ? 'Hidden' : 'Visible'}
								</button>

								<button
									onclick={setAsCover}
									disabled={selectedPhoto.id === roll.cover_photo_id}
									class="px-3 py-1.5 rounded-md text-xs border border-amber-500/30 text-amber-400 hover:border-amber-500/60 transition-colors disabled:opacity-30 disabled:cursor-default"
								>
									{selectedPhoto.id === roll.cover_photo_id ? 'Current Cover' : 'Set as Cover'}
								</button>
							</div>

							<div class="flex items-center gap-3 pt-3 border-t border-border">
								<button
									onclick={savePhoto}
									disabled={savingPhoto}
									class="px-4 py-1.5 bg-amber-600 hover:bg-amber-500 text-white rounded-md text-sm font-medium transition-colors disabled:opacity-50"
								>
									{savingPhoto ? 'Saving...' : 'Save Photo'}
								</button>
								<button
									onclick={deletePhoto}
									disabled={deletingPhoto}
									class="px-3 py-1.5 text-error/60 hover:text-error text-sm transition-colors disabled:opacity-50"
								>
									{deletingPhoto ? 'Deleting...' : 'Delete Photo'}
								</button>
							</div>
						</div>
					</div>
				</div>
			{/if}
		</div>
	</div>
{/if}
