<script lang="ts">
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';
	import type { Collection } from '$lib/types';

	let collections = $state<Collection[]>([]);
	let loading = $state(true);
	let creating = $state(false);
	let newTitle = $state('');

	$effect(() => {
		loadCollections();
	});

	async function loadCollections() {
		try {
			collections = await api.getCollections();
		} catch (e) {
			console.error('Failed to load collections:', e);
		} finally {
			loading = false;
		}
	}

	async function handleCreate() {
		if (!newTitle.trim()) return;
		creating = true;
		try {
			const coll = await api.createCollection({ title: newTitle.trim() });
			goto(`/admin/collections/${coll.id}`);
		} catch (e) {
			console.error('Failed to create collection:', e);
		} finally {
			creating = false;
		}
	}

	async function handleDelete(id: string) {
		if (!confirm('Delete this collection? Photos will not be deleted.')) return;
		try {
			await api.deleteCollection(id);
			collections = collections.filter(c => c.id !== id);
		} catch (e) {
			console.error('Failed to delete collection:', e);
		}
	}
</script>

<div class="max-w-4xl">
	<h1 class="text-2xl font-medium mb-6">Collections</h1>

	<div class="flex gap-2 mb-6">
		<input
			bind:value={newTitle}
			placeholder="New collection name"
			class="flex-1 px-3 py-2 bg-surface border border-border rounded-md text-sm focus:outline-none focus:border-accent"
			onkeydown={(e) => { if (e.key === 'Enter') handleCreate(); }}
		/>
		<button
			onclick={handleCreate}
			disabled={creating || !newTitle.trim()}
			class="px-4 py-2 bg-text text-bg rounded-md text-sm font-medium disabled:opacity-50"
		>
			Create
		</button>
	</div>

	{#if loading}
		<p class="text-text-muted">Loading...</p>
	{:else if collections.length === 0}
		<p class="text-text-muted py-8 text-center">No collections yet.</p>
	{:else}
		<div class="space-y-2">
			{#each collections as coll (coll.id)}
				<div class="flex items-center justify-between bg-surface border border-border rounded-lg p-4">
					<div>
						<a href="/admin/collections/{coll.id}" class="font-medium hover:text-accent transition-colors">
							{coll.title}
						</a>
						{#if coll.photo_count}
							<span class="text-xs text-text-muted ml-2">{coll.photo_count} photos</span>
						{/if}
					</div>
					<button
						onclick={() => handleDelete(coll.id)}
						class="text-xs text-error/60 hover:text-error"
					>
						Delete
					</button>
				</div>
			{/each}
		</div>
	{/if}
</div>
