<script lang="ts">
	import type { PageServerData } from './$types';

	let { data }: { data: PageServerData } = $props();
</script>

<svelte:head>
	<title>Collections | {data.settings?.site_title}</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 py-8">
	<h1 class="text-3xl font-medium mb-8">Collections</h1>

	{#if data.collections.length > 0}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each data.collections as collection}
				<a
					href="/collections/{collection.slug}"
					class="group block bg-surface rounded-lg overflow-hidden hover:bg-surface-hover transition-colors"
				>
					{#if collection.cover_url}
						<div class="aspect-[3/2] overflow-hidden">
							<img
								src={collection.cover_url}
								alt={collection.title}
								class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
							/>
						</div>
					{:else}
						<div class="aspect-[3/2] bg-surface flex items-center justify-center text-text-muted">
							No cover photo
						</div>
					{/if}
					<div class="p-4">
						<h2 class="font-medium">{collection.title}</h2>
						{#if collection.photo_count}
							<p class="text-sm text-text-muted mt-1">{collection.photo_count} photos</p>
						{/if}
					</div>
				</a>
			{/each}
		</div>
	{:else}
		<p class="text-center py-16 text-text-muted">No collections yet.</p>
	{/if}
</div>
