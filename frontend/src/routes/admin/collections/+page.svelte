<script lang="ts">
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';
	import type { Collection } from '$lib/types';
	import { toasts } from '$lib/stores/toast';
	import ConfirmDialog from '$lib/components/admin/ConfirmDialog.svelte';

	let collections = $state<Collection[]>([]);
	let loading = $state(true);
	let creating = $state(false);

	// Confirm dialog state
	let confirmOpen = $state(false);
	let confirmAction = $state<(() => void) | null>(null);

	$effect(() => {
		loadCollections();
	});

	async function loadCollections() {
		try {
			collections = await api.getCollections();
		} catch (e) {
			console.error('Failed to load collections:', e);
			toasts.error('Failed to load collections');
		} finally {
			loading = false;
		}
	}

	async function handleCreate() {
		creating = true;
		try {
			const coll = await api.createCollection({ title: 'Untitled Collection' });
			goto(`/admin/collections/${coll.id}`);
		} catch (e) {
			console.error('Failed to create collection:', e);
			toasts.error('Failed to create collection');
		} finally {
			creating = false;
		}
	}

	function requestDelete(id: string, e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		confirmAction = () => doDelete(id);
		confirmOpen = true;
	}

	async function doDelete(id: string) {
		confirmOpen = false;
		try {
			await api.deleteCollection(id);
			collections = collections.filter(c => c.id !== id);
			toasts.success('Collection deleted');
		} catch (e) {
			console.error('Failed to delete collection:', e);
			toasts.error('Failed to delete collection');
		}
	}
</script>

<ConfirmDialog
	open={confirmOpen}
	title="Delete Collection"
	message="Delete this collection? Photos will not be deleted."
	confirmLabel="Delete"
	onconfirm={() => { if (confirmAction) confirmAction(); }}
	oncancel={() => { confirmOpen = false; confirmAction = null; }}
/>

<div>
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-2xl font-medium">Collections</h1>
		<button
			onclick={handleCreate}
			disabled={creating}
			class="px-4 py-2 bg-amber-600 hover:bg-amber-500 text-white rounded-md text-sm font-medium transition-colors disabled:opacity-50"
		>
			{creating ? 'Creating...' : '+ New Collection'}
		</button>
	</div>

	{#if loading}
		<p class="text-text-muted">Loading...</p>
	{:else if collections.length === 0}
		<div class="text-center py-16 bg-surface border border-border rounded-lg">
			<p class="text-text-muted mb-4">No collections yet. Create one to get started.</p>
			<button
				onclick={handleCreate}
				disabled={creating}
				class="px-4 py-2 bg-amber-600 hover:bg-amber-500 text-white rounded-md text-sm font-medium transition-colors disabled:opacity-50"
			>
				{creating ? 'Creating...' : '+ New Collection'}
			</button>
		</div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each collections as coll (coll.id)}
				<a
					href="/admin/collections/{coll.id}"
					class="bg-surface border border-border rounded-lg overflow-hidden hover:border-text-muted/30 transition-colors group relative"
				>
					{#if coll.cover_url}
						<div class="aspect-[3/2] overflow-hidden">
							<img
								src={coll.cover_url}
								alt={coll.title}
								class="w-full h-full object-cover group-hover:scale-[1.02] transition-transform duration-300"
							/>
						</div>
					{:else}
						<div class="aspect-[3/2] bg-surface-hover flex items-center justify-center">
							<span class="text-text-muted/40 text-sm">No cover</span>
						</div>
					{/if}
					<div class="p-4">
						<h2 class="font-semibold text-text">{coll.title}</h2>
						<p class="text-xs text-text-muted/60 mt-1">
							{coll.photo_count ?? 0} photo{(coll.photo_count ?? 0) !== 1 ? 's' : ''}
						</p>
					</div>
					<button
						onclick={(e) => requestDelete(coll.id, e)}
						class="absolute top-2 right-2 px-2 py-1 rounded bg-black/60 text-error/70 hover:text-error text-xs opacity-0 group-hover:opacity-100 transition-opacity"
					>
						Delete
					</button>
				</a>
			{/each}
		</div>
	{/if}
</div>
