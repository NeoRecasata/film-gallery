<script lang="ts">
	import type { Photo } from '$lib/types';
	import { api } from '$lib/api';
	import MasonryGrid from '$lib/components/MasonryGrid.svelte';
	import PhotoCard from '$lib/components/PhotoCard.svelte';
	import Lightbox from '$lib/components/Lightbox.svelte';
	import InfiniteScroll from '$lib/components/InfiniteScroll.svelte';
	import type { PageServerData } from './$types';

	let { data }: { data: PageServerData } = $props();

	let photos = $state<Photo[]>(data.photosResponse.data);
	let cursor = $state<string | null>(data.photosResponse.next_cursor);
	let loading = $state(false);

	let lightboxOpen = $state(false);
	let lightboxIndex = $state(0);

	async function loadMore() {
		if (!cursor || loading) return;
		loading = true;
		try {
			const res = await api.getPhotos(cursor);
			photos = [...photos, ...res.data];
			cursor = res.next_cursor;
		} catch (e) {
			console.error('Failed to load more photos:', e);
		} finally {
			loading = false;
		}
	}

	function openLightbox(photo: Photo, index: number) {
		lightboxIndex = index;
		lightboxOpen = true;
	}
</script>

<div class="max-w-7xl mx-auto px-4 py-8">
	{#if photos.length > 0}
		<MasonryGrid {photos} onphotoclick={openLightbox}>
			{#snippet photo(props)}
				<PhotoCard photo={props.photo} />
			{/snippet}
		</MasonryGrid>

		<InfiniteScroll
			hasMore={cursor !== null}
			{loading}
			onloadmore={loadMore}
		/>
	{:else}
		<div class="text-center py-32 text-text-muted">
			<p>No photos yet.</p>
		</div>
	{/if}
</div>

{#if lightboxOpen}
	<Lightbox
		{photos}
		currentIndex={lightboxIndex}
		onclose={() => lightboxOpen = false}
	/>
{/if}
