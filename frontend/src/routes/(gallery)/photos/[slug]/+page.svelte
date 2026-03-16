<script lang="ts">
	import type { PageServerData } from './$types';

	let { data }: { data: PageServerData } = $props();
	const photo = $derived(data.photo);
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
	<img
		src={photo.urls.full || photo.urls.medium}
		alt={photo.title || ''}
		class="w-full h-auto"
		style:aspect-ratio="{photo.width} / {photo.height}"
	/>

	<div class="mt-6 space-y-2">
		{#if photo.title}
			<h1 class="text-2xl font-medium">{photo.title}</h1>
		{/if}
		{#if photo.description}
			<p class="text-text-muted">{photo.description}</p>
		{/if}
		<div class="flex gap-4 text-sm text-text-muted">
			{#if photo.film_stock}<span>{photo.film_stock}</span>{/if}
			{#if photo.camera}<span>{photo.camera}</span>{/if}
			{#if photo.lens}<span>{photo.lens}</span>{/if}
		</div>
	</div>

	<div class="mt-8">
		<a href="/" class="text-text-muted hover:text-text text-sm transition-colors">&larr; Back to gallery</a>
	</div>
</div>
