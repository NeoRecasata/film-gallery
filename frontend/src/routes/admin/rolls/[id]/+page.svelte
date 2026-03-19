<script lang="ts">
	import { page } from '$app/state';
	import { goto, beforeNavigate } from '$app/navigation';
	import { api } from '$lib/api';
	import type { Roll, Photo } from '$lib/types';
	import { toasts } from '$lib/stores/toast';
	import UploadQueue from '$lib/components/admin/UploadQueue.svelte';
	import ConfirmDialog from '$lib/components/admin/ConfirmDialog.svelte';
	import MetadataSelect from '$lib/components/admin/MetadataSelect.svelte';
	import Icon from '$lib/components/Icon.svelte';

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
	let showUploadZoneManual = $state<boolean | null>(null);

	const photos = $derived(roll?.photos || []);
	let showUploadZone = $derived.by(() => showUploadZoneManual ?? photos.length === 0);

	// Bulk select state
	let selecting = $state(false);
	let selectedIds = $state<Set<string>>(new Set());
	let bulkActing = $state(false);
	const selectedCount = $derived(selectedIds.size);
	const allSelectedFeatured = $derived(
		selectedCount > 0 && photos.filter(p => selectedIds.has(p.id)).every(p => p.featured)
	);
	const allSelectedHidden = $derived(
		selectedCount > 0 && photos.filter(p => selectedIds.has(p.id)).every(p => p.hidden)
	);
	const allSelectedVisible = $derived(
		selectedCount > 0 && photos.filter(p => selectedIds.has(p.id)).every(p => !p.hidden)
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

	const reorderDistributed = $derived.by(() => {
		const cols: Photo[][] = Array.from({ length: columnCount }, () => []);
		const heights = new Array(columnCount).fill(0);

		for (const photo of reorderPhotos) {
			const shortest = heights.indexOf(Math.min(...heights));
			cols[shortest].push(photo);
			heights[shortest] += photo.height / photo.width;
		}
		return cols;
	});

	// Flat index lookup for reorder drag operations
	const reorderIndexOf = $derived.by(() => {
		const map = new Map<string, number>();
		reorderPhotos.forEach((p, i) => map.set(p.id, i));
		return map;
	});

	$effect(() => {
		loadRoll();
	});

	beforeNavigate(({ cancel }) => {
		if (isDirty && !confirm('You have unsaved changes. Leave this page?')) {
			cancel();
		}
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
			loadSuggestions();
		} catch (e) {
			console.error('Failed to load roll:', e);
			toasts.error('Failed to load roll');
		} finally {
			loading = false;
		}
	}

	let suggestions = $state<Record<string, string[]>>({ camera: [], film_stock: [], lens: [], location: [] });

	async function loadSuggestions() {
		try {
			const allRolls = await api.getRolls();
			const fields = ['camera', 'film_stock', 'lens', 'location'] as const;
			const result: Record<string, string[]> = {};
			for (const field of fields) {
				const values = allRolls
					.map(r => r[field])
					.filter((v): v is string => v != null && v.trim() !== '');
				result[field] = [...new Set(values)].sort();
			}
			suggestions = result;
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
			loadSuggestions();
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

	async function toggleFeaturedInline(photo: Photo, e: MouseEvent) {
		e.stopPropagation();
		try {
			const updated = await api.updatePhoto(photo.id, { featured: !photo.featured } as Partial<Photo>);
			if (roll) {
				roll = { ...roll, photos: roll.photos?.map(p => p.id === updated.id ? updated : p) };
			}
			toasts.success(updated.featured ? 'Photo featured' : 'Photo unfeatured');
		} catch (e) {
			console.error('Failed to toggle featured:', e);
			toasts.error('Failed to update photo');
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

	function enterReorderMode() {
		reorderPhotos = [...photos];
		selectedPhotoId = null;
		reordering = true;
	}

	function cancelReorder() {
		reordering = false;
		reorderPhotos = [];
		dragIndex = null;
		dragOverIndex = null;
	}

	async function saveReorder() {
		if (!roll) return;
		savingReorder = true;
		try {
			const orders = reorderPhotos.map((p, i) => ({ id: p.id, sort_order: i }));
			await api.reorderRollPhotos(roll.id, orders);
			roll = { ...roll, photos: reorderPhotos.map((p, i) => ({ ...p, sort_order: i })) };
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
		// Move the dragged item to the new position
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

	// Bulk select functions
	function enterSelectMode() {
		selectedPhotoId = null;
		selectedIds = new Set();
		selecting = true;
	}

	function exitSelectMode() {
		selecting = false;
		selectedIds = new Set();
	}

	function toggleSelect(photoId: string) {
		const next = new Set(selectedIds);
		if (next.has(photoId)) {
			next.delete(photoId);
		} else {
			next.add(photoId);
		}
		selectedIds = next;
	}

	function selectAll() {
		selectedIds = new Set(photos.map(p => p.id));
	}

	async function bulkSetFeatured(featured: boolean) {
		bulkActing = true;
		try {
			await Promise.all(
				[...selectedIds].map(id => api.updatePhoto(id, { featured } as Partial<Photo>))
			);
			await loadRoll();
			toasts.success(`${selectedIds.size} photo${selectedIds.size !== 1 ? 's' : ''} ${featured ? 'featured' : 'unfeatured'}`);
			exitSelectMode();
		} catch (e) {
			console.error('Bulk feature failed:', e);
			toasts.error('Failed to update some photos');
		} finally {
			bulkActing = false;
		}
	}

	async function bulkSetHidden(hidden: boolean) {
		bulkActing = true;
		try {
			await Promise.all(
				[...selectedIds].map(id => api.updatePhoto(id, { hidden } as Partial<Photo>))
			);
			await loadRoll();
			toasts.success(`${selectedIds.size} photo${selectedIds.size !== 1 ? 's' : ''} ${hidden ? 'hidden' : 'shown'}`);
			exitSelectMode();
		} catch (e) {
			console.error('Bulk update failed:', e);
			toasts.error('Failed to update some photos');
		} finally {
			bulkActing = false;
		}
	}

	function requestBulkDelete() {
		const count = selectedIds.size;
		confirmTitle = `Delete ${count} Photo${count !== 1 ? 's' : ''}`;
		confirmMessage = `This will permanently delete ${count} photo${count !== 1 ? 's' : ''}. This action cannot be undone.`;
		confirmLabel = 'Delete';
		confirmAction = () => doBulkDelete();
		confirmOpen = true;
	}

	async function doBulkDelete() {
		confirmOpen = false;
		bulkActing = true;
		try {
			await Promise.all(
				[...selectedIds].map(id => api.deletePhoto(id))
			);
			await loadRoll();
			toasts.success(`${selectedIds.size} photo${selectedIds.size !== 1 ? 's' : ''} deleted`);
			exitSelectMode();
		} catch (e) {
			console.error('Bulk delete failed:', e);
			toasts.error('Failed to delete some photos');
		} finally {
			bulkActing = false;
		}
	}

	function handleConfirmCancel() {
		confirmOpen = false;
		confirmAction = null;
	}

	function handleConfirm() {
		if (confirmAction) confirmAction();
	}

	function handleKeyboardSave(e: KeyboardEvent) {
		if ((e.metaKey || e.ctrlKey) && e.key === 's') {
			e.preventDefault();
			if (reordering) {
				saveReorder();
			} else if (selectedPhoto) {
				savePhoto();
			} else if (isDirty) {
				saveRoll();
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
	onconfirm={handleConfirm}
	oncancel={handleConfirmCancel}
/>

{#if loading}
	<p class="text-text-muted">Loading...</p>
{:else if !roll}
	<p class="text-error">Roll not found.</p>
{:else}
	<!-- Fixed layout: top bar + content area — fill parent, no scroll on this level -->
	<div class="flex flex-col lg:h-[calc(100vh-4rem)]">
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
						<label class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Camera</label>
						<MetadataSelect bind:value={camera} options={suggestions.camera || []} placeholder="Camera" />
					</div>

					<div>
						<label class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Film Stock</label>
						<MetadataSelect bind:value={filmStock} options={suggestions.film_stock || []} placeholder="Film Stock" />
					</div>

					<div>
						<label class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Lens</label>
						<MetadataSelect bind:value={lens} options={suggestions.lens || []} placeholder="Lens" />
					</div>

					<div>
						<label class="block text-[11px] uppercase tracking-wide text-text-muted mb-1">Location</label>
						<MetadataSelect bind:value={location} options={suggestions.location || []} placeholder="Location" />
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
				<!-- Upload zone (stays at top, toggled, hidden during reorder/select) -->
				{#if !reordering && !selecting}
				<div class="flex-shrink-0">
					<UploadQueue rollId={roll.id} onuploaded={handleUploaded} showDropZone={showUploadZone} onallcomplete={() => showUploadZoneManual = false} />
				</div>
				{/if}

				<div class="mt-4 mb-3 flex items-center justify-between flex-shrink-0">
					<h2 class="text-[11px] font-semibold uppercase tracking-wider text-text-muted">
						{#if selecting}
							{selectedCount} selected
						{:else if reordering}
							Reorder Photos
						{:else}
							Photos ({photos.length})
						{/if}
					</h2>
					<div class="flex items-center gap-2">
						{#if selecting}
							<button onclick={selectAll} disabled={bulkActing} class="px-3 py-1 rounded-md text-xs font-medium bg-surface-hover text-text-muted hover:text-text transition-colors disabled:opacity-50 inline-flex items-center gap-1">
								<Icon name="list" class="w-3.5 h-3.5" /> All</button>
							{#if allSelectedHidden}
								<button onclick={() => bulkSetHidden(false)} disabled={bulkActing || selectedCount === 0} class="px-3 py-1 rounded-md text-xs font-medium bg-surface-hover text-success/70 hover:text-success transition-colors disabled:opacity-50 inline-flex items-center gap-1">
									<Icon name="eye" class="w-3.5 h-3.5" /> Show</button>
							{:else}
								<button onclick={() => bulkSetHidden(true)} disabled={bulkActing || selectedCount === 0} class="px-3 py-1 rounded-md text-xs font-medium bg-surface-hover text-text-muted hover:text-text transition-colors disabled:opacity-50 inline-flex items-center gap-1">
									<Icon name="eye-slash" class="w-3.5 h-3.5" /> Hide</button>
							{/if}
							{#if allSelectedFeatured}
								<button onclick={() => bulkSetFeatured(false)} disabled={bulkActing || selectedCount === 0} class="px-3 py-1 rounded-md text-xs font-medium bg-surface-hover text-amber-400 hover:text-amber-300 transition-colors disabled:opacity-50 inline-flex items-center gap-1">
									<Icon name="sparkle" class="w-3.5 h-3.5" /> Unfeature</button>
							{:else}
								<button onclick={() => bulkSetFeatured(true)} disabled={bulkActing || selectedCount === 0} class="px-3 py-1 rounded-md text-xs font-medium bg-surface-hover text-text-muted hover:text-amber-400 transition-colors disabled:opacity-50 inline-flex items-center gap-1">
									<Icon name="sparkle" class="w-3.5 h-3.5" /> Feature</button>
							{/if}
							<button onclick={requestBulkDelete} disabled={bulkActing || selectedCount === 0} class="px-3 py-1 rounded-md text-xs font-medium bg-surface-hover text-error/70 hover:text-error transition-colors disabled:opacity-50 inline-flex items-center gap-1">
								<Icon name="trash" class="w-3.5 h-3.5" /> Delete</button>
							<button onclick={exitSelectMode} class="px-3 py-1 rounded-md text-xs font-medium bg-surface-hover text-text-muted transition-colors inline-flex items-center gap-1">
								<Icon name="check" class="w-3.5 h-3.5" /> Done</button>
						{:else if reordering}
							<button onclick={cancelReorder} class="px-3 py-1 rounded-md text-xs font-medium bg-surface-hover text-text-muted transition-colors">Cancel</button>
							<button onclick={saveReorder} disabled={savingReorder} class="px-3 py-1 rounded-md text-xs font-medium bg-amber-600 hover:bg-amber-500 text-white transition-colors disabled:opacity-50">
								{savingReorder ? 'Saving...' : 'Done'}</button>
						{:else}
							{#if photos.length > 0}
								<button onclick={enterSelectMode} class="px-3 py-1 rounded-md text-xs font-medium bg-surface-hover text-text-muted hover:text-text transition-colors inline-flex items-center gap-1">
									<Icon name="check" class="w-3.5 h-3.5" /> Select</button>
							{/if}
							{#if photos.length > 1}
								<button onclick={enterReorderMode} class="px-3 py-1 rounded-md text-xs font-medium bg-surface-hover text-text-muted hover:text-text transition-colors inline-flex items-center gap-1">
									<Icon name="grid" class="w-3.5 h-3.5" /> Reorder</button>
							{/if}
							<button
								onclick={() => showUploadZoneManual = showUploadZone ? false : true}
								class="px-3 py-1 rounded-md text-xs font-medium transition-colors inline-flex items-center gap-1
									{showUploadZone ? 'bg-surface-hover text-text-muted' : 'bg-amber-600 hover:bg-amber-500 text-white'}"
							>
								{#if showUploadZone}
									Cancel
								{:else}
									<Icon name="plus" class="w-3.5 h-3.5" /> Add Photos
								{/if}
							</button>
						{/if}
					</div>
				</div>

				<!-- Photo grid (only this scrolls) -->
				<div class="flex-1 overflow-y-auto rounded [scrollbar-width:none] [&::-webkit-scrollbar]:hidden {selectedPhoto && !reordering && !selecting ? 'pb-[200px]' : ''}">
					{#if photos.length === 0}
					{:else if reordering}
						<!-- Masonry grid for reorder mode -->
						<div class="flex gap-2" bind:clientWidth={gridWidth}>
							{#each reorderDistributed as column}
								<div class="flex-1 flex flex-col gap-2">
									{#each column as photo (photo.id)}
										{@const flatIndex = reorderIndexOf.get(photo.id) ?? 0}
										<!-- svelte-ignore a11y_no_static_element_interactions -->
										<div
											draggable="true"
											ondragstart={() => handleDragStart(flatIndex)}
											ondragover={(e) => handleDragOver(e, flatIndex)}
											ondragend={handleDragEnd}
											class="relative rounded overflow-hidden cursor-grab active:cursor-grabbing
												{dragIndex === flatIndex ? 'opacity-40' : ''}"
										>
											<img
												src={photo.urls.thumb}
												alt={photo.title || ''}
												class="w-full h-auto block"
												style:aspect-ratio="{photo.width} / {photo.height}"
												draggable="false"
											/>
											<span class="absolute top-1 left-1 w-5 h-5 rounded bg-black/60 text-white text-[10px] font-medium flex items-center justify-center">
												{flatIndex + 1}
											</span>
											{#if photo.hidden}
												<span class="absolute top-1 right-1 px-1 py-0.5 bg-error/80 text-white text-[9px] font-semibold uppercase rounded">Hidden</span>
											{/if}
											{#if photo.featured}
												<span class="absolute bottom-1 left-1 px-1 py-0.5 bg-amber-500/80 text-white text-[9px] font-semibold uppercase rounded">Featured</span>
											{/if}
											{#if photo.id === roll.cover_photo_id}
												<span class="absolute bottom-1 right-1 px-1 py-0.5 bg-amber-500/80 text-white text-[9px] font-semibold uppercase rounded">Cover</span>
											{/if}
										</div>
									{/each}
								</div>
							{/each}
						</div>
					{:else}
						<!-- Masonry grid for normal/select mode -->
						<div class="flex gap-2" bind:clientWidth={gridWidth}>
							{#each distributed as column}
								<div class="flex-1 flex flex-col gap-2">
									{#each column as photo (photo.id)}
										<!-- svelte-ignore a11y_no_static_element_interactions -->
										<div
											onclick={() => selecting ? toggleSelect(photo.id) : selectPhoto(photo.id)}
											onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); selecting ? toggleSelect(photo.id) : selectPhoto(photo.id); } }}
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

											<!-- Hover overlay controls (hidden during select mode) -->
											{#if !selecting}
											<div class="absolute inset-0 bg-black/0 group-hover:bg-black/30 transition-colors pointer-events-none">
												<div class="absolute bottom-0 left-0 right-0 p-1.5 flex items-center justify-end gap-1 opacity-100 lg:opacity-0 group-hover:opacity-100 transition-opacity pointer-events-auto">
													<!-- Toggle hidden -->
													<button
														onclick={(e) => togglePhotoHiddenInline(photo, e)}
														title={photo.hidden ? 'Show photo' : 'Hide photo'}
														class="w-9 h-9 lg:w-7 lg:h-7 rounded bg-black/60 hover:bg-black/80 text-white flex items-center justify-center text-base lg:text-sm transition-colors"
													>
														<Icon name={photo.hidden ? 'eye-slash' : 'eye'} />
													</button>
													<button
														onclick={(e) => setAsCoverInline(photo, e)}
														disabled={photo.id === roll.cover_photo_id}
														title={photo.id === roll.cover_photo_id ? 'Current cover' : 'Set as cover'}
														class="w-9 h-9 lg:w-7 lg:h-7 rounded bg-black/60 flex items-center justify-center text-base lg:text-sm transition-colors
															{photo.id === roll.cover_photo_id ? 'text-yellow-400 cursor-default' : 'hover:bg-black/80 text-white'}"
													>
														<Icon name="star" />
													</button>
													<button
														onclick={(e) => toggleFeaturedInline(photo, e)}
														title={photo.featured ? 'Unfeature' : 'Feature'}
														class="w-9 h-9 lg:w-7 lg:h-7 rounded bg-black/60 hover:bg-black/80 flex items-center justify-center text-base lg:text-sm transition-colors
															{photo.featured ? 'text-amber-400' : 'text-white'}"
													>
														<Icon name="sparkle" />
													</button>
													<button
														onclick={(e) => requestDeletePhotoInline(photo, e)}
														title="Delete photo"
														class="w-9 h-9 lg:w-7 lg:h-7 rounded bg-black/60 hover:bg-error/80 text-white flex items-center justify-center text-base lg:text-sm transition-colors"
													>
														<Icon name="trash" />
													</button>
												</div>
											</div>
											{/if}

											<!-- Selection overlays -->
											{#if selecting}
												<div class="absolute top-1.5 left-1.5 w-5 h-5 rounded border-2 flex items-center justify-center pointer-events-none
													{selectedIds.has(photo.id) ? 'bg-amber-500 border-amber-500' : 'border-white/60 bg-black/30'}">
													{#if selectedIds.has(photo.id)}
														<Icon name="check" class="w-3.5 h-3.5 text-white" />
													{/if}
												</div>
											{:else if selectedPhotoId === photo.id}
												<div class="absolute inset-0 rounded border-2 border-amber-500 pointer-events-none"></div>
											{/if}

											<!-- Badges -->
											{#if photo.hidden}
												<span class="absolute top-1 {selecting ? 'left-8' : 'left-1'} px-1 py-0.5 bg-error/80 text-white text-[9px] font-semibold uppercase rounded">
													Hidden
												</span>
											{/if}
											{#if photo.featured}
												<span class="absolute bottom-1 left-1 px-1 py-0.5 bg-amber-500/80 text-white text-[9px] font-semibold uppercase rounded">
													Featured
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

<!-- Floating photo editor - fixed to bottom (hidden during reorder/select, desktop only) -->
{#if selectedPhoto && roll && !reordering && !selecting}
	<div class="hidden lg:block fixed bottom-4 left-[calc(140px+50vw)] -translate-x-1/2 z-50 bg-surface border border-border rounded-lg shadow-[0_-4px_24px_rgba(0,0,0,0.4)] w-[min(56rem,calc(100vw-280px-3rem))]">
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

