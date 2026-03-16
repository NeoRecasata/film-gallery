<script lang="ts">
	import type { Photo } from '$lib/types';

	import type { Snippet } from 'svelte';

	let { photos, onphotoclick, photo: photoSnippet }: {
		photos: Photo[];
		onphotoclick: (photo: Photo, index: number) => void;
		photo: Snippet<[{ photo: Photo; index: number }]>;
	} = $props();

	let columns = $state(3);
	let containerWidth = $state(0);

	$effect(() => {
		if (containerWidth < 640) columns = 1;
		else if (containerWidth < 1024) columns = 2;
		else columns = 3;
	});

	const distributed = $derived.by(() => {
		const cols: Photo[][] = Array.from({ length: columns }, () => []);
		const heights = new Array(columns).fill(0);

		for (const photo of photos) {
			const shortest = heights.indexOf(Math.min(...heights));
			cols[shortest].push(photo);
			heights[shortest] += photo.height / photo.width;
		}
		return cols;
	});
</script>

<div
	class="flex gap-2 sm:gap-3"
	bind:clientWidth={containerWidth}
>
	{#each distributed as column, colIdx}
		<div class="flex-1 flex flex-col gap-2 sm:gap-3">
			{#each column as photo, photoIdx}
				{@const globalIndex = photos.indexOf(photo)}
				<button
					class="block w-full cursor-pointer"
					onclick={() => onphotoclick(photo, globalIndex)}
				>
					{@render photoSnippet({ photo, index: globalIndex })}
				</button>
			{/each}
		</div>
	{/each}
</div>
