<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>Rolls | {data.settings?.site_title}</title>
	{#if data.rolls.length > 0 && data.rolls[0].cover_url}
		<meta property="og:image" content={data.rolls[0].cover_url} />
	{/if}
</svelte:head>

<div class="max-w-7xl mx-auto px-4 py-8">
	<div class="flex items-center justify-between mb-8">
		<h1 class="text-3xl font-medium">Rolls</h1>
		<a href="/rolls/browse" class="text-sm text-text-muted hover:text-text transition-colors">Browse all photos &rarr;</a>
	</div>

	{#if data.rolls.length > 0}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each data.rolls as roll}
				<a
					href="/rolls/{roll.slug}"
					class="group block bg-surface rounded-lg overflow-hidden hover:bg-surface-hover transition-colors"
				>
					{#if roll.cover_url}
						<div class="aspect-[3/2] overflow-hidden">
							<img
								src={roll.cover_url}
								alt={roll.title}
								class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
							/>
						</div>
					{:else}
						<div class="aspect-[3/2] bg-surface flex items-center justify-center text-text-muted">
							No cover photo
						</div>
					{/if}
					<div class="p-4">
						<h2 class="font-medium">{roll.title}</h2>
						{#if roll.camera || roll.film_stock}
							<p class="text-sm text-text-muted mt-1">
								{[roll.camera, roll.film_stock].filter(Boolean).join(' · ')}
							</p>
						{/if}
						{#if roll.photo_count}
							<p class="text-xs text-text-muted mt-1">{roll.photo_count} photos</p>
						{/if}
					</div>
				</a>
			{/each}
		</div>
	{:else}
		<p class="text-center py-16 text-text-muted">No rolls yet.</p>
	{/if}
</div>
