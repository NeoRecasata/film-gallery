<script lang="ts">
	import type { Photo } from '$lib/types';
	import MasonryGrid from '$lib/components/MasonryGrid.svelte';
	import PhotoCard from '$lib/components/PhotoCard.svelte';
	import Lightbox from '$lib/components/Lightbox.svelte';
	import type { PageServerData } from './$types';

	let { data }: { data: PageServerData } = $props();
	const roll = $derived(data.roll);

	let lightboxOpen = $state(false);
	let lightboxIndex = $state(0);

	function openLightbox(photo: Photo, index: number) {
		lightboxIndex = index;
		lightboxOpen = true;
	}

	const metaItems = $derived(
		[roll.camera, roll.film_stock, roll.lens, roll.location, roll.shot_at ? formatDate(roll.shot_at) : null]
			.filter(Boolean)
	);

	function formatDate(dateStr: string): string {
		try {
			return new Date(dateStr).toLocaleDateString('en-US', {
				year: 'numeric',
				month: 'long',
			});
		} catch {
			return dateStr;
		}
	}
</script>

<svelte:head>
	<title>{roll.title} | {data.settings?.site_title}</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 py-8">
	<div class="mb-8">
		<a href="/rolls" class="text-text-muted hover:text-text text-sm transition-colors">&larr; Rolls</a>
		<h1 class="text-3xl font-medium mt-2">{roll.title}</h1>
		{#if roll.description}
			<p class="text-text-muted mt-2">{roll.description}</p>
		{/if}
		{#if metaItems.length > 0}
			<p class="text-sm text-text-muted mt-2">{metaItems.join(' · ')}</p>
		{/if}
	</div>

	{#if roll.photos && roll.photos.length > 0}
		<MasonryGrid photos={roll.photos} onphotoclick={openLightbox}>
			{#snippet photo(props)}
				<PhotoCard photo={props.photo} />
			{/snippet}
		</MasonryGrid>
	{:else}
		<p class="text-center py-16 text-text-muted">No photos in this roll.</p>
	{/if}
</div>

{#if lightboxOpen && roll.photos}
	<Lightbox
		photos={roll.photos}
		currentIndex={lightboxIndex}
		onclose={() => lightboxOpen = false}
	/>
{/if}
