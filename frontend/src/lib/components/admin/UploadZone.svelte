<script lang="ts">
	import { api } from '$lib/api';
	import type { Photo } from '$lib/types';

	let { onuploaded }: { onuploaded: (photo: Photo) => void } = $props();

	type UploadItem = {
		file: File;
		status: 'pending' | 'uploading' | 'done' | 'failed';
		progress: number;
		photo?: Photo;
		error?: string;
	};

	let items = $state<UploadItem[]>([]);
	let dragging = $state(false);
	let fileInput: HTMLInputElement | undefined = $state();

	const MAX_CONCURRENT = 3;

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
		const newItems: UploadItem[] = imageFiles.map(f => ({
			file: f,
			status: 'pending',
			progress: 0
		}));
		items = [...items, ...newItems];
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
		items = items; // trigger reactivity

		try {
			const photo = await api.uploadPhoto(item.file);
			item.status = 'done';
			item.photo = photo;
			onuploaded(photo);
		} catch (e) {
			item.status = 'failed';
			item.error = e instanceof Error ? e.message : 'Upload failed';
		}

		items = items; // trigger reactivity
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
</script>

<div>
	<!-- Drop zone -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="border-2 border-dashed rounded-lg p-8 text-center transition-colors cursor-pointer
			{dragging ? 'border-accent bg-accent/5' : 'border-border hover:border-text-muted'}"
		ondragover={(e) => { e.preventDefault(); dragging = true; }}
		ondragleave={() => dragging = false}
		ondrop={handleDrop}
		onclick={() => fileInput?.click()}
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

	<!-- Upload items -->
	{#if items.length > 0}
		<div class="mt-4 space-y-2">
			{#if items.some(i => i.status === 'done')}
				<button onclick={clearDone} class="text-xs text-text-muted hover:text-text">
					Clear completed
				</button>
			{/if}

			{#each items as item}
				<div class="flex items-center gap-3 bg-surface rounded-md p-3">
					{#if item.photo?.urls.thumb}
						<img src={item.photo.urls.thumb} alt="" class="w-10 h-10 rounded object-cover" />
					{:else}
						<div class="w-10 h-10 rounded bg-surface-hover flex items-center justify-center text-xs text-text-muted">
							{item.file.name.slice(-4)}
						</div>
					{/if}

					<div class="flex-1 min-w-0">
						<p class="text-sm truncate">{item.file.name}</p>
						<p class="text-xs text-text-muted">
							{(item.file.size / 1024 / 1024).toFixed(1)} MB
						</p>
					</div>

					<div class="text-sm">
						{#if item.status === 'uploading'}
							<div class="w-5 h-5 border-2 border-text-muted border-t-transparent rounded-full animate-spin"></div>
						{:else if item.status === 'done'}
							<span class="text-success">Done</span>
						{:else if item.status === 'failed'}
							<div class="flex items-center gap-2">
								<span class="text-error text-xs">{item.error}</span>
								<button onclick={() => retryItem(item)} class="text-text-muted hover:text-text text-xs underline">
									Retry
								</button>
							</div>
						{:else}
							<span class="text-text-muted">Pending</span>
						{/if}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
