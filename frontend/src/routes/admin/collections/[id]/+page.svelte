<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api';
	import type { Collection, Photo } from '$lib/types';
	import { toasts } from '$lib/stores/toast';
	import ConfirmDialog from '$lib/components/admin/ConfirmDialog.svelte';

	let collectionId = $derived(page.params.id as string);
	let collection = $state<Collection | null>(null);
	let allPhotos = $state<Photo[]>([]);
	let loading = $state(true);
	let saving = $state(false);
	let deleting = $state(false);

	// Form state
	let title = $state('');
	let description = $state('');

	// Saved values for dirty tracking
	let savedTitle = $state('');
	let savedDescription = $state('');

	const isDirty = $derived(
		title !== savedTitle || description !== savedDescription
	);

	// Photo picker state
	let showPicker = $state(false);
	let selectedIds = $state<string[]>([]);
	let savedIds = $state<string[]>([]);
	let savingPhotos = $state(false);
	const pickerChanged = $derived(
		JSON.stringify(selectedIds) !== JSON.stringify(savedIds)
	);

	// Reorder state
	let reordering = $state(false);
	let reorderPhotos = $state<Photo[]>([]);
	let savingReorder = $state(false);
	let dragIndex = $state<number | null>(null);
	let dragOverIndex = $state<number | null>(null);

	// Confirm dialog state
	let confirmOpen = $state(false);
	let confirmTitle = $state('');
	let confirmMessage = $state('');
	let confirmLabel = $state('Delete');
	let confirmAction = $state<(() => void) | null>(null);

	const photos = $derived(collection?.photos || []);

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
		loadCollection();
	});

	function syncFormFromCollection(data: Collection) {
		title = data.title;
		description = data.description || '';
		savedTitle = title;
		savedDescription = description;
		const ids = (data.photos || []).map(p => p.id);
		selectedIds = [...ids];
		savedIds = [...ids];
	}

	async function loadCollection() {
		loading = true;
		try {
			const data = await api.getAdminCollection(collectionId);
			collection = data;
			syncFormFromCollection(data);
		} catch (e) {
			console.error('Failed to load collection:', e);
			toasts.error('Failed to load collection');
		} finally {
			loading = false;
		}
	}

	async function loadAllPhotos() {
		try {
			const res = await api.getAdminPhotos();
			allPhotos = res.data;
		} catch (e) {
			console.error('Failed to load photos:', e);
			toasts.error('Failed to load photos');
		}
	}

	async function saveCollection() {
		if (!collection) return;
		saving = true;
		try {
			const updated = await api.updateCollection(collection.id, {
				title: title.trim(),
				description: description.trim() || null
			} as Partial<Collection>);
			collection = { ...collection, ...updated };
			syncFormFromCollection(collection);
			toasts.success('Collection saved');
		} catch (e) {
			console.error('Failed to save:', e);
			toasts.error('Failed to save collection');
		} finally {
			saving = false;
		}
	}

	function requestDeleteCollection() {
		confirmTitle = 'Delete Collection';
		confirmMessage = 'This will permanently delete this collection. Photos will not be deleted.';
		confirmLabel = 'Delete Collection';
		confirmAction = () => doDeleteCollection();
		confirmOpen = true;
	}

	async function doDeleteCollection() {
		if (!collection) return;
		confirmOpen = false;
		deleting = true;
		try {
			await api.deleteCollection(collection.id);
			toasts.success('Collection deleted');
			goto('/admin/collections');
		} catch (e) {
			console.error('Failed to delete:', e);
			toasts.error('Failed to delete collection');
			deleting = false;
		}
	}

	// Photo picker
	function openPicker() {
		loadAllPhotos();
		showPicker = true;
	}

	function closePicker() {
		showPicker = false;
		// Reset selection if unsaved
		if (pickerChanged) {
			selectedIds = [...savedIds];
		}
	}

	function togglePhoto(photoId: string) {
		const idx = selectedIds.indexOf(photoId);
		if (idx >= 0) {
			selectedIds = selectedIds.filter(id => id !== photoId);
		} else {
			selectedIds = [...selectedIds, photoId];
		}
	}

	async function savePickerSelection() {
		if (!collection) return;
		savingPhotos = true;
		try {
			const photoList = selectedIds.map((id, i) => ({ photo_id: id, sort_order: i }));
			await api.setCollectionPhotos(collection.id, photoList);
			savedIds = [...selectedIds];
			showPicker = false;
			await loadCollection();
			toasts.success('Photos updated');
		} catch (e) {
			console.error('Failed to save photos:', e);
			toasts.error('Failed to save photos');
		} finally {
			savingPhotos = false;
		}
	}

	// Remove photo from collection
	async function removePhoto(photo: Photo, e: MouseEvent) {
		e.stopPropagation();
		if (!collection) return;
		const remaining = photos
			.filter(p => p.id !== photo.id)
			.map((p, i) => ({ photo_id: p.id, sort_order: i }));
		try {
			await api.setCollectionPhotos(collection.id, remaining);
			await loadCollection();
			toasts.success('Photo removed');
		} catch (e) {
			console.error('Failed to remove photo:', e);
			toasts.error('Failed to remove photo');
		}
	}

	// Set cover photo
	async function setCoverInline(photo: Photo, e: MouseEvent) {
		e.stopPropagation();
		if (!collection) return;
		try {
			const updated = await api.updateCollection(collection.id, { cover_photo: photo.id } as Partial<Collection>);
			collection = { ...collection, ...updated };
			toasts.success('Cover photo updated');
		} catch (e) {
			console.error('Failed to set cover:', e);
			toasts.error('Failed to set cover photo');
		}
	}

	// Reorder
	function enterReorderMode() {
		reorderPhotos = [...photos];
		reordering = true;
	}

	function cancelReorder() {
		reordering = false;
		reorderPhotos = [];
		dragIndex = null;
		dragOverIndex = null;
	}

	async function saveReorder() {
		if (!collection) return;
		savingReorder = true;
		try {
			const orders = reorderPhotos.map((p, i) => ({ photo_id: p.id, sort_order: i }));
			await api.setCollectionPhotos(collection.id, orders);
			await loadCollection();
			toasts.success('Photo order saved');
			reordering = false;
			reorderPhotos = [];
		} catch (e) {
			console.error('Failed to save order:', e);
			toasts.error('Failed to save photo order');
		} finally {
			savingReorder = false;
		}
	}

	function handleDragStart(index: number) {
		dragIndex = index;
	}

	function handleDragOver(e: DragEvent, index: number) {
		e.preventDefault();
		if (dragIndex === null || dragIndex === index) return;
		const updated = [...reorderPhotos];
		const [moved] = updated.splice(dragIndex, 1);
		updated.splice(index, 0, moved);
		reorderPhotos = updated;
		dragIndex = index;
		dragOverIndex = index;
	}

	function handleDragEnd() {
		dragIndex = null;
		dragOverIndex = null;
	}

	function handleKeyboardSave(e: KeyboardEvent) {
		if ((e.metaKey || e.ctrlKey) && e.key === 's') {
			e.preventDefault();
			if (reordering) {
				saveReorder();
			} else if (showPicker && pickerChanged) {
				savePickerSelection();
			} else if (isDirty) {
				saveCollection();
			}
		}
	}
</script>

<svelte:window onkeydown={handleKeyboardSave} />

<ConfirmDialog
	open={confirmOpen}
	title={confirmTitle}
	message={confirmMessage}
	confirmLabel={confirmLabel}
	onconfirm={() => { if (confirmAction) confirmAction(); }}
	oncancel={() => { confirmOpen = false; confirmAction = null; }}
/>

{#if loading}
	<p class="text-text-muted">Loading...</p>
{:else if !collection}
	<p class="text-error">Collection not found.</p>
{:else}
	<div class="flex flex-col" style="height: calc(100vh - 4rem);">
		<!-- Top bar -->
		<div class="flex items-center justify-between pb-4 flex-shrink-0 border-b border-border">
			<div class="flex items-center gap-2 text-sm">
				<a href="/admin/collections" class="text-text-muted hover:text-text transition-colors">&larr; Collections</a>
				<span class="text-text-muted/40">/</span>
				<span class="text-text">{collection.title}</span>
			</div>
			<div class="flex items-center gap-3">
				<button
					onclick={requestDeleteCollection}
					disabled={deleting}
					class="px-3 py-1.5 border border-error/30 text-error/70 hover:text-error hover:border-error/60 rounded-md text-sm transition-colors disabled:opacity-50"
				>
					{deleting ? 'Deleting...' : 'Delete Collection'}
				</button>
				<button
					onclick={saveCollection}
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
			<!-- Left sidebar - metadata -->
			<div class="w-full lg:w-[280px] flex-shrink-0 overflow-y-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
				<div class="bg-surface border border-border rounded-lg p-5 space-y-4">
					<h2 class="text-[11px] font-semibold uppercase tracking-wider text-text-muted">Metadata</h2>
					<div>
						<label for="coll-title" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Title</label>
						<input
							id="coll-title"
							bind:value={title}
							class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent"
						/>
					</div>
					<div>
						<label for="coll-desc" class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Description</label>
						<textarea
							id="coll-desc"
							bind:value={description}
							rows="3"
							class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent resize-none"
						></textarea>
					</div>
				</div>
			</div>

			<!-- Right area - photos -->
			<div class="flex-1 min-w-0 flex flex-col min-h-0">
				<div class="mt-0 mb-3 flex items-center justify-between flex-shrink-0">
					<h2 class="text-[11px] font-semibold uppercase tracking-wider text-text-muted">
						{#if reordering}
							Reorder Photos
						{:else if showPicker}
							Add Photos
							{#if selectedIds.length > 0}
								<span class="font-normal ml-1">({selectedIds.length} selected)</span>
							{/if}
						{:else}
							Photos ({photos.length})
						{/if}
					</h2>
					<div class="flex items-center gap-2">
						{#if reordering}
							<button onclick={cancelReorder} class="px-3 py-1 rounded-md text-xs font-medium bg-surface-hover text-text-muted transition-colors">Cancel</button>
							<button onclick={saveReorder} disabled={savingReorder} class="px-3 py-1 rounded-md text-xs font-medium bg-amber-600 hover:bg-amber-500 text-white transition-colors disabled:opacity-50">
								{savingReorder ? 'Saving...' : 'Done'}
							</button>
						{:else if showPicker}
							<button onclick={closePicker} class="px-3 py-1 rounded-md text-xs font-medium bg-surface-hover text-text-muted transition-colors">Cancel</button>
							{#if pickerChanged}
								<button onclick={savePickerSelection} disabled={savingPhotos} class="px-3 py-1 rounded-md text-xs font-medium bg-amber-600 hover:bg-amber-500 text-white transition-colors disabled:opacity-50">
									{savingPhotos ? 'Saving...' : 'Save'}
								</button>
							{/if}
						{:else}
							{#if photos.length > 1}
								<button onclick={enterReorderMode} class="px-3 py-1 rounded-md text-xs font-medium bg-surface-hover text-text-muted hover:text-text transition-colors">Reorder</button>
							{/if}
							<button
								onclick={openPicker}
								class="px-3 py-1 rounded-md text-xs font-medium bg-amber-600 hover:bg-amber-500 text-white transition-colors"
							>
								+ Add Photos
							</button>
						{/if}
					</div>
				</div>

				<!-- Photo grid -->
				<div class="flex-1 overflow-y-auto rounded [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
					{#if showPicker}
						<!-- Photo picker: Instagram-style selection grid -->
						<div class="grid grid-cols-3 sm:grid-cols-5 md:grid-cols-6 gap-2">
							{#each allPhotos as photo (photo.id)}
								{@const selIndex = selectedIds.indexOf(photo.id)}
								{@const isSelected = selIndex >= 0}
								<button
									onclick={() => togglePhoto(photo.id)}
									class="rounded overflow-hidden text-left relative group transition-opacity
										{isSelected ? '' : 'opacity-50 hover:opacity-80'}"
								>
									<div class="aspect-square">
										<img src={photo.urls.thumb} alt={photo.title || ''} class="w-full h-full object-cover" />
									</div>
									{#if isSelected}
										<div class="absolute inset-0 rounded border-2 border-amber-500 pointer-events-none"></div>
										<span class="absolute top-1.5 left-1.5 w-5 h-5 rounded-full bg-amber-500 text-white text-[10px] font-bold flex items-center justify-center shadow">
											{selIndex + 1}
										</span>
									{/if}
									{#if photo.roll_title}
										<p class="px-1.5 py-1 text-[10px] text-text-muted truncate">{photo.roll_title}</p>
									{/if}
								</button>
							{/each}
						</div>
					{:else if reordering}
						<!-- Flat grid for reorder -->
						<div class="grid grid-cols-4 gap-2">
							{#each reorderPhotos as photo, index (photo.id)}
								<!-- svelte-ignore a11y_no_static_element_interactions -->
								<div
									draggable="true"
									ondragstart={() => handleDragStart(index)}
									ondragover={(e) => handleDragOver(e, index)}
									ondragend={handleDragEnd}
									class="relative rounded overflow-hidden cursor-grab active:cursor-grabbing
										{dragIndex === index ? 'opacity-40' : ''}
										{dragOverIndex === index && dragIndex !== index ? 'ring-2 ring-amber-500' : ''}"
								>
									<img src={photo.urls.thumb} alt={photo.title || ''} class="w-full aspect-square object-cover block" draggable="false" />
									<span class="absolute top-1 left-1 w-5 h-5 rounded bg-black/60 text-white text-[10px] font-medium flex items-center justify-center">{index + 1}</span>
								</div>
							{/each}
						</div>
					{:else if photos.length === 0}
						<p class="text-text-muted text-sm text-center py-12">No photos in this collection yet.</p>
					{:else}
						<!-- Masonry grid -->
						<div class="flex gap-2" bind:clientWidth={gridWidth}>
							{#each distributed as column}
								<div class="flex-1 flex flex-col gap-2">
									{#each column as photo (photo.id)}
										<div class="block w-full rounded overflow-hidden relative group">
											<img
												src={photo.urls.thumb}
												alt={photo.title || ''}
												class="w-full h-auto block"
												style:aspect-ratio="{photo.width} / {photo.height}"
											/>
											<!-- Hover overlay -->
											<div class="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors pointer-events-none">
												<div class="absolute bottom-0 left-0 right-0 p-1.5 flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-auto">
													<!-- Set as cover -->
													<button
														onclick={(e) => setCoverInline(photo, e)}
														disabled={photo.id === collection.cover_photo}
														title={photo.id === collection.cover_photo ? 'Current cover' : 'Set as cover'}
														class="w-7 h-7 rounded bg-black/60 flex items-center justify-center text-sm transition-colors
															{photo.id === collection.cover_photo ? 'text-yellow-400 cursor-default' : 'hover:bg-black/80 text-white'}"
													>
														<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
															<path fill-rule="evenodd" d="M10.868 2.884c-.321-.772-1.415-.772-1.736 0l-1.83 4.401-4.753.381c-.833.067-1.171 1.107-.536 1.651l3.62 3.102-1.106 4.637c-.194.813.691 1.456 1.405 1.02L10 15.591l4.069 2.485c.713.436 1.598-.207 1.404-1.02l-1.106-4.637 3.62-3.102c.635-.544.297-1.584-.536-1.65l-4.752-.382-1.831-4.401z" clip-rule="evenodd" />
														</svg>
													</button>
													<!-- Remove from collection -->
													<button
														onclick={(e) => removePhoto(photo, e)}
														title="Remove from collection"
														class="w-7 h-7 rounded bg-black/60 hover:bg-error/80 text-white flex items-center justify-center text-sm transition-colors"
													>
														<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4">
															<path d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
														</svg>
													</button>
												</div>
											</div>
											<!-- Badges -->
											{#if photo.id === collection.cover_photo}
												<span class="absolute top-1 right-1 px-1 py-0.5 bg-amber-500/80 text-white text-[9px] font-semibold uppercase rounded">Cover</span>
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
