<script lang="ts">
	import type { Photo } from '$lib/types';
	import MasonryGrid from '$lib/components/MasonryGrid.svelte';
	import PhotoCard from '$lib/components/PhotoCard.svelte';
	import Lightbox from '$lib/components/Lightbox.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	const rolls = $derived(data.rolls);

	// Flatten all photos for lightbox navigation
	const allPhotos = $derived(rolls.flatMap(r => r.photos || []));

	let lightboxOpen = $state(false);
	let lightboxIndex = $state(0);

	// Cumulative photo offset per roll for global lightbox indexing
	const photoOffsets = $derived(
		rolls.map((_r, i) => rolls.slice(0, i).reduce((sum, prev) => sum + (prev.photos?.length || 0), 0))
	);

	function openLightbox(_photo: Photo, _index: number, globalIndex: number) {
		lightboxIndex = globalIndex;
		lightboxOpen = true;
	}
</script>

<svelte:head>
	<title>Browse All Photos | {data.settings?.site_title}</title>
</svelte:head>

<div class="max-w-7xl mx-auto px-4 py-8">
	<div class="mb-8">
		<a href="/rolls" class="text-text-muted hover:text-text text-sm transition-colors">&larr; Rolls</a>
		<h1 class="text-3xl font-medium mt-2">Browse All Photos</h1>
	</div>

	{#if rolls.length > 0}
		{#each rolls as roll, rollIndex}
			{#if roll.photos && roll.photos.length > 0}
				{@const meta = [roll.camera, roll.film_stock, roll.lens, roll.location].filter(Boolean)}
				<section class="mb-12">
					<a href="/rolls/{roll.slug}" class="group">
						<h2 class="text-xl font-medium group-hover:text-white transition-colors">{roll.title}</h2>
						{#if meta.length > 0}
							<p class="text-sm text-text-muted mt-1">{meta.join(' · ')}</p>
						{/if}
					</a>
					<div class="mt-4">
						<MasonryGrid photos={roll.photos} onphotoclick={(photo, index) => openLightbox(photo, index, photoOffsets[rollIndex] + index)}>
							{#snippet photo(props)}
								<PhotoCard photo={props.photo} showOverlay />
							{/snippet}
						</MasonryGrid>
					</div>
				</section>
			{/if}
		{/each}
	{:else}
		<p class="text-center py-16 text-text-muted">No photos yet.</p>
	{/if}
</div>

{#if lightboxOpen && allPhotos.length > 0}
	<Lightbox
		photos={allPhotos}
		currentIndex={lightboxIndex}
		onclose={() => lightboxOpen = false}
	/>
{/if}
