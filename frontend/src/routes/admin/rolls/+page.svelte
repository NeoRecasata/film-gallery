<script lang="ts">
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';
	import type { Roll } from '$lib/types';

	let rolls = $state<Roll[]>([]);
	let loading = $state(true);
	let creating = $state(false);

	$effect(() => {
		loadRolls();
	});

	async function loadRolls() {
		try {
			rolls = await api.getRolls();
		} catch (e) {
			console.error('Failed to load rolls:', e);
		} finally {
			loading = false;
		}
	}

	function getLastMetadata(): { camera?: string; film_stock?: string; lens?: string; location?: string } {
		try {
			const stored = localStorage.getItem('film-gallery-last-roll-metadata');
			if (stored) {
				const parsed = JSON.parse(stored);
				const result: Record<string, string> = {};
				if (parsed.camera) result.camera = parsed.camera;
				if (parsed.film_stock) result.film_stock = parsed.film_stock;
				if (parsed.lens) result.lens = parsed.lens;
				if (parsed.location) result.location = parsed.location;
				return result;
			}
		} catch { /* ignore */ }
		return {};
	}

	async function handleCreate() {
		creating = true;
		try {
			const lastMeta = getLastMetadata();
			const roll = await api.createRoll({ title: 'Untitled Roll', ...lastMeta });
			goto(`/admin/rolls/${roll.id}`);
		} catch (e) {
			console.error('Failed to create roll:', e);
		} finally {
			creating = false;
		}
	}
</script>

<div>
	<div class="flex items-center justify-between mb-6">
		<h1 class="text-2xl font-medium">Rolls</h1>
		<button
			onclick={handleCreate}
			disabled={creating}
			class="px-4 py-2 bg-amber-600 hover:bg-amber-500 text-white rounded-md text-sm font-medium transition-colors disabled:opacity-50"
		>
			{creating ? 'Creating...' : '+ New Roll'}
		</button>
	</div>

	{#if loading}
		<p class="text-text-muted">Loading...</p>
	{:else if rolls.length === 0}
		<p class="text-text-muted py-12 text-center">No rolls yet. Create one to get started.</p>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each rolls as roll (roll.id)}
				<a
					href="/admin/rolls/{roll.id}"
					class="bg-surface border border-border rounded-lg overflow-hidden hover:border-text-muted/30 transition-colors group"
				>
					{#if roll.cover_url}
						<div class="aspect-[3/2] overflow-hidden">
							<img
								src={roll.cover_url}
								alt={roll.title}
								class="w-full h-full object-cover group-hover:scale-[1.02] transition-transform duration-300"
							/>
						</div>
					{:else}
						<div class="aspect-[3/2] bg-surface-hover flex items-center justify-center">
							<span class="text-text-muted/40 text-sm">No cover</span>
						</div>
					{/if}
					<div class="p-4">
						<div class="flex items-center gap-2 mb-1">
							<h2 class="font-semibold text-text">{roll.title}</h2>
							<span class="px-1.5 py-0.5 rounded text-[10px] font-medium uppercase tracking-wide
								{roll.published ? 'bg-success/15 text-success' : 'bg-amber-500/15 text-amber-400'}">
								{roll.published ? 'Published' : 'Draft'}
							</span>
						</div>
						{#if roll.camera || roll.film_stock}
							<p class="text-sm text-text-muted">
								{[roll.camera, roll.film_stock].filter(Boolean).join(' \u00b7 ')}
							</p>
						{/if}
						<p class="text-xs text-text-muted/60 mt-1">
							{roll.photo_count ?? 0} photo{(roll.photo_count ?? 0) !== 1 ? 's' : ''}
						</p>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>
