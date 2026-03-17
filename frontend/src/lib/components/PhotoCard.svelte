<script lang="ts">
	import type { Photo } from '$lib/types';
	import { blurhashToDataURL } from '$lib/blurhash';
	import { browser } from '$app/environment';

	let { photo, showOverlay = false }: { photo: Photo; showOverlay?: boolean } = $props();

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

	const metaLine = $derived.by(() => {
		const parts: string[] = [];
		if (photo.camera) parts.push(photo.camera);
		if (photo.film_stock) parts.push(photo.film_stock);
		if (photo.lens) parts.push(photo.lens);
		return parts.join(' · ');
	});

	$effect(() => {
		if (imgEl && imgEl.complete) loaded = true;
	});
</script>

<div
	class="relative overflow-hidden rounded-sm bg-surface group"
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

	{#if showOverlay && (photo.title || metaLine || photo.location)}
		<div class="absolute inset-0 bg-gradient-to-t from-black/70 via-transparent to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none flex items-end">
			<div class="p-3 w-full">
				{#if photo.title}
					<p class="text-white text-sm font-medium truncate">{photo.title}</p>
				{/if}
				{#if metaLine}
					<p class="text-white/70 text-xs truncate mt-0.5">{metaLine}</p>
				{/if}
				{#if photo.location}
					<p class="text-white/50 text-xs truncate mt-0.5">{photo.location}</p>
				{/if}
			</div>
		</div>
	{/if}
</div>
