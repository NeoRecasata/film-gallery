<script lang="ts">
	import type { Photo } from '$lib/types';
	import { blurhashToDataURL } from '$lib/blurhash';
	import { browser } from '$app/environment';

	let { photo }: { photo: Photo } = $props();

	let loaded = $state(false);
	let imgEl: HTMLImageElement | undefined = $state();

	const aspectRatio = $derived(`${photo.width} / ${photo.height}`);
	const placeholder = $derived(
		browser && photo.blur_hash ? blurhashToDataURL(photo.blur_hash, 32, 32) : null
	);

	const srcset = $derived.by(() => {
		const parts: string[] = [];
		if (photo.urls.thumb) parts.push(`${photo.urls.thumb} 400w`);
		if (photo.urls.medium) parts.push(`${photo.urls.medium} 1200w`);
		if (photo.urls.full) parts.push(`${photo.urls.full} 2400w`);
		return parts.join(', ');
	});

	$effect(() => {
		if (imgEl && imgEl.complete) loaded = true;
	});
</script>

<div
	class="relative overflow-hidden rounded-sm bg-surface"
	style:aspect-ratio={aspectRatio}
>
	{#if placeholder && !loaded}
		<img
			src={placeholder}
			alt=""
			class="absolute inset-0 w-full h-full object-cover"
		/>
	{/if}

	<img
		bind:this={imgEl}
		src={photo.urls.medium || photo.urls.thumb}
		srcset={srcset}
		sizes="(max-width: 640px) 100vw, (max-width: 1024px) 50vw, 33vw"
		alt={photo.title || ''}
		loading="lazy"
		class="w-full h-full object-cover transition-opacity duration-300"
		class:opacity-0={!loaded}
		onload={() => loaded = true}
	/>
</div>
