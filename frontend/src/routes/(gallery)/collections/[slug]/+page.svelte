<script lang="ts">
	import type { Photo } from '$lib/types';
	import MasonryGrid from '$lib/components/MasonryGrid.svelte';
	import PhotoCard from '$lib/components/PhotoCard.svelte';
	import Lightbox from '$lib/components/Lightbox.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	const collection = $derived(data.collection);

	let lightboxOpen = $state(false);
	let lightboxIndex = $state(0);

	function openLightbox(photo: Photo, index: number) {
		lightboxIndex = index;
		lightboxOpen = true;
	}
</script>

<svelte:head>
	<title>{collection.title} | {data.settings?.site_title}</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 py-8">
	<div class="mb-8">
		<a href="/collections" class="text-text-muted hover:text-text text-sm transition-colors">&larr; Collections</a>
		<h1 class="text-3xl font-medium mt-2">{collection.title}</h1>
		{#if collection.description}
			<p class="text-text-muted mt-2">{collection.description}</p>
		{/if}
	</div>

	{#if collection.photos && collection.photos.length > 0}
		<MasonryGrid photos={collection.photos} onphotoclick={openLightbox}>
			{#snippet photo(props)}
				<PhotoCard photo={props.photo} showOverlay />
			{/snippet}
		</MasonryGrid>
	{:else}
		<p class="text-center py-16 text-text-muted">No photos in this collection.</p>
	{/if}
</div>

{#if lightboxOpen && collection.photos}
	<Lightbox
		photos={collection.photos}
		currentIndex={lightboxIndex}
		onclose={() => lightboxOpen = false}
	/>
{/if}
