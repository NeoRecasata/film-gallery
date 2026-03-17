<script lang="ts">
	import { api } from '$lib/api';
	import type { Photo } from '$lib/types';
	import { toasts } from '$lib/stores/toast';

	let { rollId, onuploaded }: { rollId: string; onuploaded: (photo: Photo) => void } = $props();

	type UploadItem = {
		file: File;
		status: 'pending' | 'uploading' | 'done' | 'failed';
		progress: number;
		photo?: Photo;
		error?: string;
	};

	let items = $state<UploadItem[]>([]);
	let dragging = $state(false);
	let collapsed = $state(false);
	let fileInput: HTMLInputElement | undefined = $state();

	const MAX_CONCURRENT = 3;

	const hasItems = $derived(items.length > 0);
	const activeCount = $derived(items.filter(i => i.status === 'uploading' || i.status === 'pending').length);
	const doneCount = $derived(items.filter(i => i.status === 'done').length);
	const failedCount = $derived(items.filter(i => i.status === 'failed').length);
	const isComplete = $derived(hasItems && activeCount === 0);

	// Auto-clear completed uploads after 60 seconds
	let autoClearTimer: ReturnType<typeof setTimeout> | null = null;

	$effect(() => {
		if (isComplete && doneCount > 0 && failedCount === 0) {
			autoClearTimer = setTimeout(() => {
				items = items.filter(i => i.status !== 'done');
			}, 60000);
		}
		return () => {
			if (autoClearTimer) clearTimeout(autoClearTimer);
		};
	});

	const headerText = $derived.by(() => {
		if (!hasItems) return '';
		if (activeCount > 0) return `Uploading ${activeCount} file${activeCount !== 1 ? 's' : ''}...`;
		if (failedCount > 0) return `${doneCount} uploaded, ${failedCount} failed`;
		return `Upload complete (${doneCount})`;
	});

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		dragging = false;
		const files = Array.from(e.dataTransfer?.files || []);
		addFiles(files);
	}

	function handleFileSelect(e: Event) {
		const input = e.target as HTMLInputElement;
		const files = Array.from(input.files || []);
		addFiles(files);
		input.value = '';
	}

	function addFiles(files: File[]) {
		const imageFiles = files.filter(f =>
			['image/jpeg', 'image/png', 'image/tiff'].includes(f.type)
		);
		if (imageFiles.length === 0) return;
		const newItems: UploadItem[] = imageFiles.map(f => ({
			file: f,
			status: 'pending' as const,
			progress: 0
		}));
		items = [...items, ...newItems];
		collapsed = false;
		processQueue();
	}

	function processQueue() {
		const uploading = items.filter(i => i.status === 'uploading').length;
		const pending = items.filter(i => i.status === 'pending');
		const toStart = pending.slice(0, MAX_CONCURRENT - uploading);

		for (const item of toStart) {
			uploadFile(item);
		}
	}

	async function uploadFile(item: UploadItem) {
		item.status = 'uploading';
		items = items;

		try {
			const result = await api.uploadRollPhotos(rollId, [item.file]);
			if (result.uploaded.length > 0) {
				item.status = 'done';
				item.photo = result.uploaded[0];
				onuploaded(result.uploaded[0]);
			} else if (result.failed.length > 0) {
				throw new Error(result.failed[0].error);
			}
		} catch (e) {
			item.status = 'failed';
			item.error = e instanceof Error ? e.message : 'Upload failed';
			toasts.error(`Failed to upload ${item.file.name}`);
		}

		items = items;
		processQueue();
	}

	function retryItem(item: UploadItem) {
		item.status = 'pending';
		item.error = undefined;
		items = items;
		processQueue();
	}

	function clearDone() {
		items = items.filter(i => i.status !== 'done');
	}

	function clearAll() {
		items = items.filter(i => i.status === 'uploading');
	}
</script>

<!-- Drop zone in main content -->
<div>
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div
		class="border-2 border-dashed rounded-lg p-8 text-center transition-colors cursor-pointer
			{dragging ? 'border-accent bg-accent/5' : 'border-border hover:border-text-muted'}"
		ondragover={(e) => { e.preventDefault(); dragging = true; }}
		ondragleave={() => dragging = false}
		ondrop={handleDrop}
		onclick={() => fileInput?.click()}
		role="button"
		tabindex="0"
		onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); fileInput?.click(); } }}
	>
		<p class="text-text-muted">
			{dragging ? 'Drop files here' : 'Drag photos here or click to browse'}
		</p>
		<p class="text-xs text-text-muted mt-1">JPEG, PNG, TIFF (max 100MB each)</p>
	</div>

	<input
		bind:this={fileInput}
		type="file"
		accept="image/jpeg,image/png,image/tiff"
		multiple
		class="hidden"
		onchange={handleFileSelect}
	/>
</div>

<!-- Floating upload queue widget -->
{#if hasItems}
	<div class="fixed bottom-4 right-4 z-[90] w-[360px] bg-surface border border-border rounded-lg shadow-2xl overflow-hidden">
		<!-- Header -->
		<div class="w-full flex items-center justify-between px-4 py-3 bg-surface-hover">
			<button
				onclick={() => collapsed = !collapsed}
				class="flex items-center gap-2 hover:opacity-80 transition-opacity"
			>
				{#if activeCount > 0}
					<div class="w-4 h-4 border-2 border-accent border-t-transparent rounded-full animate-spin flex-shrink-0"></div>
				{:else if failedCount > 0}
					<span class="text-error text-sm flex-shrink-0">&times;</span>
				{:else}
					<span class="text-success text-sm flex-shrink-0">&#10003;</span>
				{/if}
				<span class="text-sm font-medium text-text">{headerText}</span>
				<span class="text-text-muted text-xs transition-transform {collapsed ? '' : 'rotate-180'}">
					&#9660;
				</span>
			</button>
			{#if isComplete}
				<button
					onclick={clearAll}
					class="text-xs text-text-muted hover:text-text transition-colors"
				>
					Clear
				</button>
			{/if}
		</div>

		<!-- File list -->
		{#if !collapsed}
			<div class="max-h-[260px] overflow-y-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden divide-y divide-border">
				{#each items as item}
					<div class="flex items-center gap-3 px-4 py-2.5">
						{#if item.photo?.urls.thumb}
							<img src={item.photo.urls.thumb} alt="" class="w-8 h-8 rounded object-cover flex-shrink-0" />
						{:else}
							<div class="w-8 h-8 rounded bg-bg flex items-center justify-center text-[10px] text-text-muted flex-shrink-0">
								IMG
							</div>
						{/if}

						<div class="flex-1 min-w-0">
							<p class="text-xs truncate text-text">{item.file.name}</p>
							<p class="text-[10px] text-text-muted">
								{(item.file.size / 1024 / 1024).toFixed(1)} MB
							</p>
						</div>

						<div class="flex-shrink-0">
							{#if item.status === 'uploading'}
								<div class="w-4 h-4 border-2 border-text-muted border-t-transparent rounded-full animate-spin"></div>
							{:else if item.status === 'done'}
								<span class="text-success text-sm">&#10003;</span>
							{:else if item.status === 'failed'}
								<div class="flex items-center gap-1.5">
									<span class="text-error text-sm">&times;</span>
									<button onclick={() => retryItem(item)} class="text-[10px] text-text-muted hover:text-text underline">
										Retry
									</button>
								</div>
							{:else}
								<span class="text-text-muted text-[10px]">Pending</span>
							{/if}
						</div>
					</div>
				{/each}
			</div>

			{#if items.some(i => i.status === 'done')}
				<div class="px-4 py-2 border-t border-border">
					<button onclick={clearDone} class="text-[10px] text-text-muted hover:text-text transition-colors">
						Clear completed
					</button>
				</div>
			{/if}
		{/if}
	</div>
{/if}
