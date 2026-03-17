<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import type { Roll, Photo } from '$lib/types';
	import { toasts } from '$lib/stores/toast';
	import UploadQueue from '$lib/components/admin/UploadQueue.svelte';
	import ConfirmDialog from '$lib/components/admin/ConfirmDialog.svelte';

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

	// Saved/original values for dirty tracking
	let savedTitle = $state('');
	let savedSlug = $state('');
	let savedDescription = $state('');
	let savedCamera = $state('');
	let savedFilmStock = $state('');
	let savedLens = $state('');
	let savedLocation = $state('');
	let savedShotAt = $state('');
	let savedPublished = $state(false);

	// Dirty tracking
	const isDirty = $derived(
		title !== savedTitle ||
		slug !== savedSlug ||
		description !== savedDescription ||
		camera !== savedCamera ||
		filmStock !== savedFilmStock ||
		lens !== savedLens ||
		location !== savedLocation ||
		shotAt !== savedShotAt ||
		published !== savedPublished
	);

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

	// Confirm dialog state
	let confirmOpen = $state(false);
	let confirmTitle = $state('');
	let confirmMessage = $state('');
	let confirmLabel = $state('Delete');
	let confirmAction = $state<(() => void) | null>(null);

	const photos = $derived(roll?.photos || []);
	const selectedPhoto = $derived(photos.find(p => p.id === selectedPhotoId) || null);

	// Masonry columns
	let gridWidth = $state(0);
	const columnCount = $derived(gridWidth < 500 ? 2 : gridWidth < 800 ? 3 : 4);

	const distributed = $derived.by(() => {
		const cols: Photo[][] = Array.from({ length: columnCount }, () => []);
		const heights = new Array(columnCount).fill(0);

		for (const photo of photos) {
			const shortest = heights.indexOf(Math.min(...heights));
			cols[shortest].push(photo);
			heights[shortest] += photo.height / photo.width;
		}
		return cols;
	});

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

	function syncFormFromRoll(data: Roll) {
		title = data.title;
		slug = data.slug;
		description = data.description || '';
		camera = data.camera || '';
		filmStock = data.film_stock || '';
		lens = data.lens || '';
		location = data.location || '';
		shotAt = data.shot_at ? data.shot_at.split('T')[0] : '';
		published = data.published;
		// Save copies for dirty tracking
		savedTitle = title;
		savedSlug = slug;
		savedDescription = description;
		savedCamera = camera;
		savedFilmStock = filmStock;
		savedLens = lens;
		savedLocation = location;
		savedShotAt = shotAt;
		savedPublished = published;
	}

	async function loadRoll() {
		loading = true;
		try {
			const data = await api.getRoll(rollId);
			roll = data;
			syncFormFromRoll(data);
		} catch (e) {
			console.error('Failed to load roll:', e);
			toasts.error('Failed to load roll');
		} finally {
			loading = false;
		}
	}

	function storeLastMetadata() {
		try {
			localStorage.setItem('film-gallery-last-roll-metadata', JSON.stringify({
				camera: camera.trim() || null,
				film_stock: filmStock.trim() || null,
				lens: lens.trim() || null,
				location: location.trim() || null
			}));
		} catch { /* ignore */ }
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
			syncFormFromRoll(roll);
			storeLastMetadata();
			toasts.success('Roll saved');
		} catch (e) {
			console.error('Failed to save roll:', e);
			toasts.error('Failed to save roll');
		} finally {
			saving = false;
		}
	}

	function requestDeleteRoll() {
		confirmTitle = 'Delete Roll';
		confirmMessage = 'This will permanently delete this roll and all its photos. This action cannot be undone.';
		confirmLabel = 'Delete Roll';
		confirmAction = () => doDeleteRoll();
		confirmOpen = true;
	}

	async function doDeleteRoll() {
		if (!roll) return;
		confirmOpen = false;
		deleting = true;
		try {
			await api.deleteRoll(roll.id);
			toasts.success('Roll deleted');
			goto('/admin/rolls');
		} catch (e) {
			console.error('Failed to delete roll:', e);
			toasts.error('Failed to delete roll');
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
			toasts.success('Photo saved');
		} catch (e) {
			console.error('Failed to save photo:', e);
			toasts.error('Failed to save photo');
		} finally {
			savingPhoto = false;
		}
	}

	async function togglePhotoHiddenInline(photo: Photo, e: MouseEvent) {
		e.stopPropagation();
		try {
			const updated = await api.updatePhoto(photo.id, { hidden: !photo.hidden } as Partial<Photo>);
			if (roll) {
				roll = { ...roll, photos: roll.photos?.map(p => p.id === updated.id ? updated : p) };
			}
			// If editing this photo, sync
			if (selectedPhotoId === photo.id) {
				photoHidden = updated.hidden;
			}
			toasts.success(updated.hidden ? 'Photo hidden' : 'Photo visible');
		} catch (e) {
			console.error('Failed to toggle hidden:', e);
			toasts.error('Failed to update photo');
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
			toasts.success(updated.hidden ? 'Photo hidden' : 'Photo visible');
		} catch (e) {
			console.error('Failed to toggle hidden:', e);
			toasts.error('Failed to update photo');
			photoHidden = !photoHidden;
		}
	}

	async function setAsCoverInline(photo: Photo, e: MouseEvent) {
		e.stopPropagation();
		if (!roll) return;
		try {
			const updated = await api.updateRoll(roll.id, { cover_photo_id: photo.id });
			roll = { ...roll, ...updated };
			toasts.success('Cover photo updated');
		} catch (e) {
			console.error('Failed to set cover:', e);
			toasts.error('Failed to set cover photo');
		}
	}

	async function setAsCover() {
		if (!roll || !selectedPhotoId) return;
		try {
			const updated = await api.updateRoll(roll.id, { cover_photo_id: selectedPhotoId });
			roll = { ...roll, ...updated };
			toasts.success('Cover photo updated');
		} catch (e) {
			console.error('Failed to set cover:', e);
			toasts.error('Failed to set cover photo');
		}
	}

	function requestDeletePhotoInline(photo: Photo, e: MouseEvent) {
		e.stopPropagation();
		confirmTitle = 'Delete Photo';
		confirmMessage = 'This will permanently delete this photo. This action cannot be undone.';
		confirmLabel = 'Delete Photo';
		confirmAction = () => doDeletePhoto(photo.id);
		confirmOpen = true;
	}

	function requestDeletePhoto() {
		if (!selectedPhoto) return;
		confirmTitle = 'Delete Photo';
		confirmMessage = 'This will permanently delete this photo. This action cannot be undone.';
		confirmLabel = 'Delete Photo';
		confirmAction = () => doDeletePhoto(selectedPhoto!.id);
		confirmOpen = true;
	}

	async function doDeletePhoto(photoId: string) {
		confirmOpen = false;
		deletingPhoto = true;
		try {
			await api.deletePhoto(photoId);
			if (roll) {
				roll = {
					...roll,
					photos: roll.photos?.filter(p => p.id !== photoId),
					photo_count: Math.max(0, (roll.photo_count ?? 1) - 1)
				};
			}
			if (selectedPhotoId === photoId) {
				selectedPhotoId = null;
			}
			toasts.success('Photo deleted');
		} catch (e) {
			console.error('Failed to delete photo:', e);
			toasts.error('Failed to delete photo');
		} finally {
			deletingPhoto = false;
		}
	}

	function handleConfirmCancel() {
		confirmOpen = false;
		confirmAction = null;
	}

	function handleConfirm() {
		if (confirmAction) confirmAction();
	}
</script>

<ConfirmDialog
	open={confirmOpen}
	title={confirmTitle}
	message={confirmMessage}
	confirmLabel={confirmLabel}
	onconfirm={handleConfirm}
	oncancel={handleConfirmCancel}
/>

{#if loading}
	<p class="text-text-muted">Loading...</p>
{:else if !roll}
	<p class="text-error">Roll not found.</p>
{:else}
	<!-- Fixed layout: top bar + content area — fill parent, no scroll on this level -->
	<div class="flex flex-col" style="height: calc(100vh - 4rem);">
		<!-- Top bar -->
		<div class="flex items-center justify-between pb-4 flex-shrink-0 border-b border-border">
			<div class="flex items-center gap-2 text-sm">
				<a href="/admin/rolls" class="text-text-muted hover:text-text transition-colors">&larr; Rolls</a>
				<span class="text-text-muted/40">/</span>
				<span class="text-text">{roll.title}</span>
			</div>
			<div class="flex items-center gap-3">
				<button
					onclick={requestDeleteRoll}
					disabled={deleting}
					class="px-3 py-1.5 border border-error/30 text-error/70 hover:text-error hover:border-error/60 rounded-md text-sm transition-colors disabled:opacity-50"
				>
					{deleting ? 'Deleting...' : 'Delete Roll'}
				</button>
				<button
					onclick={saveRoll}
					disabled={saving || !isDirty}
					class="px-4 py-1.5 rounded-md text-sm font-medium transition-colors disabled:opacity-50
						{isDirty ? 'bg-amber-600 hover:bg-amber-500 text-white' : 'bg-surface-hover text-text-muted cursor-default'}"
				>
					{saving ? 'Saving...' : 'Save'}
				</button>
			</div>
		</div>

		<!-- Content area: sidebar + photos -->
		<div class="flex flex-col lg:flex-row gap-6 flex-1 min-h-0 overflow-hidden pt-6">
			<!-- Left sidebar - metadata (scrolls independently, hidden scrollbar) -->
			<div class="w-full lg:w-[280px] flex-shrink-0 overflow-y-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
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
			<div class="flex-1 min-w-0 flex flex-col min-h-0">
				<!-- Upload zone (stays at top) -->
				<div class="flex-shrink-0">
					<UploadQueue rollId={roll.id} onuploaded={handleUploaded} />
				</div>

				<h2 class="mt-4 text-[11px] font-semibold uppercase tracking-wider text-text-muted mb-3 flex-shrink-0">
					Photos ({photos.length})
				</h2>

				<!-- Photo grid (only this scrolls) -->
				<div class="flex-1 overflow-y-auto rounded [scrollbar-width:none] [&::-webkit-scrollbar]:hidden {selectedPhoto ? 'pb-[200px]' : ''}">
					{#if photos.length === 0}
						<p class="text-text-muted text-sm py-8 text-center">No photos yet. Upload some above.</p>
					{:else}
						<div class="flex gap-2" bind:clientWidth={gridWidth}>
							{#each distributed as column}
								<div class="flex-1 flex flex-col gap-2">
									{#each column as photo (photo.id)}
										<!-- svelte-ignore a11y_no_static_element_interactions -->
										<div
											onclick={() => selectPhoto(photo.id)}
											onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); selectPhoto(photo.id); } }}
											role="button"
											tabindex="0"
											class="block w-full rounded overflow-hidden relative group cursor-pointer"
										>
											<img
												src={photo.urls.thumb}
												alt={photo.title || ''}
												class="w-full h-auto block"
												style:aspect-ratio="{photo.width} / {photo.height}"
											/>

											<!-- Hover overlay controls -->
											<div class="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors pointer-events-none">
												<div class="absolute bottom-0 left-0 right-0 p-1.5 flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-auto">
													<!-- Toggle hidden -->
													<button
														onclick={(e) => togglePhotoHiddenInline(photo, e)}
														title={photo.hidden ? 'Show photo' : 'Hide photo'}
														class="w-7 h-7 rounded bg-black/60 hover:bg-black/80 text-white flex items-center justify-center text-sm transition-colors"
													>
														{#if photo.hidden}
															<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
																<path fill-rule="evenodd" d="M3.28 2.22a.75.75 0 00-1.06 1.06l14.5 14.5a.75.75 0 101.06-1.06l-1.745-1.745a10.029 10.029 0 003.3-4.38 1.651 1.651 0 000-1.185A10.004 10.004 0 009.999 3a9.956 9.956 0 00-4.744 1.194L3.28 2.22zM7.752 6.69l1.092 1.092a2.5 2.5 0 013.374 3.373l1.092 1.092a4 4 0 00-5.558-5.558z" clip-rule="evenodd" />
																<path d="M10.748 13.93l2.523 2.523a9.987 9.987 0 01-3.27.547c-4.258 0-7.894-2.66-9.337-6.41a1.651 1.651 0 010-1.186A10.007 10.007 0 012.839 6.02L6.07 9.252a4 4 0 004.678 4.678z" />
															</svg>
														{:else}
															<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
																<path d="M10 12.5a2.5 2.5 0 100-5 2.5 2.5 0 000 5z" />
																<path fill-rule="evenodd" d="M.664 10.59a1.651 1.651 0 010-1.186A10.004 10.004 0 0110 3c4.257 0 7.893 2.66 9.336 6.41.147.381.146.804 0 1.186A10.004 10.004 0 0110 17c-4.257 0-7.893-2.66-9.336-6.41zM14 10a4 4 0 11-8 0 4 4 0 018 0z" clip-rule="evenodd" />
															</svg>
														{/if}
													</button>

													<!-- Set as cover -->
													<button
														onclick={(e) => setAsCoverInline(photo, e)}
														disabled={photo.id === roll.cover_photo_id}
														title={photo.id === roll.cover_photo_id ? 'Current cover' : 'Set as cover'}
														class="w-7 h-7 rounded bg-black/60 flex items-center justify-center text-sm transition-colors
															{photo.id === roll.cover_photo_id ? 'text-yellow-400 cursor-default' : 'hover:bg-black/80 text-white'}"
													>
														<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
															<path fill-rule="evenodd" d="M10.868 2.884c-.321-.772-1.415-.772-1.736 0l-1.83 4.401-4.753.381c-.833.067-1.171 1.107-.536 1.651l3.62 3.102-1.106 4.637c-.194.813.691 1.456 1.405 1.02L10 15.591l4.069 2.485c.713.436 1.598-.207 1.404-1.02l-1.106-4.637 3.62-3.102c.635-.544.297-1.584-.536-1.65l-4.752-.382-1.831-4.401z" clip-rule="evenodd" />
														</svg>
													</button>

													<!-- Delete -->
													<button
														onclick={(e) => requestDeletePhotoInline(photo, e)}
														title="Delete photo"
														class="w-7 h-7 rounded bg-black/60 hover:bg-error/80 text-white flex items-center justify-center text-sm transition-colors"
													>
														<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
															<path fill-rule="evenodd" d="M8.75 1A2.75 2.75 0 006 3.75v.443c-.795.077-1.584.176-2.365.298a.75.75 0 10.23 1.482l.149-.022 1.005 11.36c.09 1.017.946 1.789 1.966 1.789h6.03c1.02 0 1.876-.772 1.966-1.789l1.005-11.36.149.022a.75.75 0 10.23-1.482A41.03 41.03 0 0014 4.193V3.75A2.75 2.75 0 0011.25 1h-2.5zM10 4c.84 0 1.673.025 2.5.075V3.75c0-.69-.56-1.25-1.25-1.25h-2.5c-.69 0-1.25.56-1.25 1.25v.325C8.327 4.025 9.16 4 10 4zM8.58 7.72a.75.75 0 00-1.5.06l.3 7.5a.75.75 0 101.5-.06l-.3-7.5zm4.34.06a.75.75 0 10-1.5-.06l-.3 7.5a.75.75 0 101.5.06l.3-7.5z" clip-rule="evenodd" />
														</svg>
													</button>
												</div>
											</div>

											<!-- Selection border overlay -->
											{#if selectedPhotoId === photo.id}
												<div class="absolute inset-0 rounded border-2 border-amber-500 pointer-events-none"></div>
											{/if}

											<!-- Badges -->
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
										</div>
									{/each}
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		</div>
	</div>
{/if}

<!-- Floating photo editor - fixed to bottom -->
{#if selectedPhoto && roll}
	<div class="fixed bottom-4 right-10 z-50 bg-surface border border-border rounded-lg shadow-[0_-4px_24px_rgba(0,0,0,0.4)]" style="left: calc(14rem + 2rem + 280px + 1.5rem + 2rem);">
		<div class="flex gap-5 p-4 max-w-full">
			<!-- Preview (left) -->
			<div class="w-[160px] flex-shrink-0">
				<img
					src={selectedPhoto.urls.medium || selectedPhoto.urls.thumb}
					alt={selectedPhoto.title || ''}
					class="w-full rounded object-contain max-h-[140px]"
				/>
				<div class="flex items-center gap-2 mt-2">
					<button
						onclick={togglePhotoHidden}
						class="px-2 py-0.5 rounded text-[11px] border transition-colors
							{photoHidden ? 'border-error/30 text-error/70' : 'border-success/30 text-success'}"
					>
						{photoHidden ? 'Hidden' : 'Visible'}
					</button>
					<button
						onclick={setAsCover}
						disabled={selectedPhoto.id === roll.cover_photo_id}
						class="px-2 py-0.5 rounded text-[11px] border border-yellow-500/30 text-yellow-400 hover:border-yellow-500/60 transition-colors disabled:opacity-30 disabled:cursor-default"
					>
						{selectedPhoto.id === roll.cover_photo_id ? 'Cover' : 'Set Cover'}
					</button>
				</div>
			</div>

			<!-- Fields -->
			<div class="flex-1 min-w-0 space-y-2">
				<div class="grid grid-cols-4 gap-2">
					<div>
						<label for="photo-title" class="block text-[10px] uppercase tracking-wide text-text-muted mb-0.5">Title</label>
						<input id="photo-title" bind:value={photoTitle} placeholder="Optional title"
							class="w-full px-2 py-1.5 bg-bg border border-border rounded text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40" />
					</div>
					<div>
						<label for="photo-camera" class="block text-[10px] uppercase tracking-wide text-text-muted mb-0.5">Camera</label>
						<input id="photo-camera" bind:value={photoCamera} placeholder={camera || 'Camera'}
							class="w-full px-2 py-1.5 bg-bg border border-border rounded text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40" />
					</div>
					<div>
						<label for="photo-film" class="block text-[10px] uppercase tracking-wide text-text-muted mb-0.5">Film Stock</label>
						<input id="photo-film" bind:value={photoFilmStock} placeholder={filmStock || 'Film Stock'}
							class="w-full px-2 py-1.5 bg-bg border border-border rounded text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40" />
					</div>
					<div>
						<label for="photo-lens" class="block text-[10px] uppercase tracking-wide text-text-muted mb-0.5">Lens</label>
						<input id="photo-lens" bind:value={photoLens} placeholder={lens || 'Lens'}
							class="w-full px-2 py-1.5 bg-bg border border-border rounded text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40" />
					</div>
				</div>
				<div class="grid grid-cols-4 gap-2">
					<div>
						<label for="photo-desc" class="block text-[10px] uppercase tracking-wide text-text-muted mb-0.5">Description</label>
						<input id="photo-desc" bind:value={photoDescription} placeholder="Description"
							class="w-full px-2 py-1.5 bg-bg border border-border rounded text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40" />
					</div>
					<div>
						<label for="photo-location" class="block text-[10px] uppercase tracking-wide text-text-muted mb-0.5">Location</label>
						<input id="photo-location" bind:value={photoLocation} placeholder={location || 'Location'}
							class="w-full px-2 py-1.5 bg-bg border border-border rounded text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40" />
					</div>
					<div>
						<label for="photo-date" class="block text-[10px] uppercase tracking-wide text-text-muted mb-0.5">Date Taken</label>
						<input id="photo-date" type="date" bind:value={photoTakenAt}
							class="w-full px-2 py-1.5 bg-bg border border-border rounded text-sm focus:outline-none focus:border-accent" />
					</div>
					<div class="flex items-end gap-2">
						<button onclick={savePhoto} disabled={savingPhoto}
							class="px-3 py-1.5 bg-amber-600 hover:bg-amber-500 text-white rounded text-sm font-medium transition-colors disabled:opacity-50">
							{savingPhoto ? 'Saving...' : 'Save'}
						</button>
						<button onclick={requestDeletePhoto} disabled={deletingPhoto}
							class="px-2 py-1.5 text-error/60 hover:text-error text-sm transition-colors disabled:opacity-50">
							Delete
						</button>
						<button onclick={() => selectedPhotoId = null}
							class="px-2 py-1.5 text-text-muted hover:text-text text-sm transition-colors ml-auto">
							&times;
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}
