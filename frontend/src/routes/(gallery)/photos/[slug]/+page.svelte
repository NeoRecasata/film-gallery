<script lang="ts">
	import type { PageServerData } from './$types';

	let { data }: { data: PageServerData } = $props();
	const photo = $derived(data.photo);
	const metaItems = $derived(
		[photo.camera, photo.film_stock, photo.lens, photo.location].filter(Boolean)
	);
</script>

<svelte:head>
	<title>{photo.title || 'Photo'} | {data.settings?.site_title}</title>
	{#if photo.urls.medium}
		<meta property="og:image" content={photo.urls.medium} />
	{/if}
	<meta property="og:type" content="article" />
	{#if photo.description}
		<meta property="og:description" content={photo.description} />
	{/if}
</svelte:head>

<div class="max-w-5xl mx-auto px-4 py-8">
	<!-- Back link -->
	<div class="mb-6">
		{#if photo.roll_slug}
			<a href="/rolls/{photo.roll_slug}" class="text-text-muted hover:text-text text-sm transition-colors">&larr; {photo.roll_title || 'Back to roll'}</a>
		{:else}
			<a href="/" class="text-text-muted hover:text-text text-sm transition-colors">&larr; Back to gallery</a>
		{/if}
	</div>

	<!-- Photo with prev/next -->
	<div class="relative group">
		<img
			src={photo.urls.full || photo.urls.medium}
			alt={photo.title || ''}
			class="w-full h-auto"
			style:aspect-ratio="{photo.width} / {photo.height}"
		/>

		{#if photo.prev_slug}
			<a
				href="/photos/{photo.prev_slug}"
				class="absolute left-4 top-1/2 -translate-y-1/2 text-white/40 hover:text-white text-4xl p-2 opacity-0 group-hover:opacity-100 transition-opacity"
				aria-label="Previous photo"
			>&lsaquo;</a>
		{/if}
		{#if photo.next_slug}
			<a
				href="/photos/{photo.next_slug}"
				class="absolute right-4 top-1/2 -translate-y-1/2 text-white/40 hover:text-white text-4xl p-2 opacity-0 group-hover:opacity-100 transition-opacity"
				aria-label="Next photo"
			>&rsaquo;</a>
		{/if}
	</div>

	<!-- Metadata -->
	<div class="mt-6 space-y-2">
		{#if photo.title}
			<h1 class="text-2xl font-medium">{photo.title}</h1>
		{/if}
		{#if photo.description}
			<p class="text-text-muted">{photo.description}</p>
		{/if}
		{#if metaItems.length > 0}
			<p class="text-sm text-text-muted">{metaItems.join(' · ')}</p>
		{/if}
	</div>

	<!-- Prev/next text links -->
	<div class="flex items-center justify-between mt-8 text-sm">
		{#if photo.prev_slug}
			<a href="/photos/{photo.prev_slug}" class="text-text-muted hover:text-text transition-colors">&larr; Previous</a>
		{:else}
			<span></span>
		{/if}
		{#if photo.next_slug}
			<a href="/photos/{photo.next_slug}" class="text-text-muted hover:text-text transition-colors">Next &rarr;</a>
		{/if}
	</div>
</div>
