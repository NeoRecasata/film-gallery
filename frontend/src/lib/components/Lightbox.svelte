<script lang="ts">
	import { fade } from 'svelte/transition';
	import type { Photo } from '$lib/types';

	let {
		photos,
		currentIndex = 0,
		onclose
	}: {
		photos: Photo[];
		currentIndex: number;
		onclose: () => void;
	} = $props();

	let index = $state(currentIndex);
	let showMeta = $state(false);
	let touchStartX = $state(0);

	const photo = $derived(photos[index]);

	// Lock body scroll while lightbox is open
	$effect(() => {
		document.body.style.overflow = 'hidden';
		return () => { document.body.style.overflow = ''; };
	});

	function next() {
		if (index < photos.length - 1) index++;
	}

	function prev() {
		if (index > 0) index--;
	}

	function handleKeydown(e: KeyboardEvent) {
		switch (e.key) {
			case 'ArrowRight': next(); break;
			case 'ArrowLeft': prev(); break;
			case 'Escape': onclose(); break;
			case 'i': showMeta = !showMeta; break;
		}
	}

	function handleTouchStart(e: TouchEvent) {
		touchStartX = e.touches[0].clientX;
	}

	let copied = $state(false);

	function copyLink() {
		const url = `${window.location.origin}/photos/${photo.slug}`;
		navigator.clipboard.writeText(url);
		copied = true;
		setTimeout(() => copied = false, 2000);
	}

	function handleTouchEnd(e: TouchEvent) {
		const diff = e.changedTouches[0].clientX - touchStartX;
		if (Math.abs(diff) > 50) {
			if (diff > 0) prev();
			else next();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
	transition:fade={{ duration: 150 }}
	class="fixed inset-0 z-[100] bg-black flex items-center justify-center"
	role="dialog"
	aria-modal="true"
	ontouchstart={handleTouchStart}
	ontouchend={handleTouchEnd}
>
	<!-- Close button -->
	<button
		class="absolute top-4 right-4 z-10 text-white/60 hover:text-white text-2xl p-2"
		onclick={onclose}
		aria-label="Close"
	>&times;</button>

	<!-- Nav arrows -->
	{#if index > 0}
		<button
			class="absolute left-4 top-1/2 -translate-y-1/2 z-10 text-white/40 hover:text-white text-4xl p-2"
			onclick={prev}
			aria-label="Previous"
		>&lsaquo;</button>
	{/if}
	{#if index < photos.length - 1}
		<button
			class="absolute right-4 top-1/2 -translate-y-1/2 z-10 text-white/40 hover:text-white text-4xl p-2"
			onclick={next}
			aria-label="Next"
		>&rsaquo;</button>
	{/if}

	<!-- Photo -->
	{#if photo}
		<img
			src={photo.urls.full || photo.urls.medium}
			alt={photo.title || ''}
			class="max-w-full max-h-full object-contain select-none"
		/>
	{/if}

	<!-- Bottom bar: counter + metadata + info toggle -->
	<div class="absolute bottom-0 left-0 right-0 z-10">
		{#if showMeta && photo}
			<div class="bg-gradient-to-t from-black/80 to-transparent p-6 pt-16">
				{#if photo.title}
					<h2 class="text-white text-lg font-medium">{photo.title}</h2>
				{/if}
				{#if photo.description}
					<p class="text-white/70 text-sm mt-1">{photo.description}</p>
				{/if}
				<div class="flex items-center gap-4 mt-2 text-white/50 text-xs">
					<span class="text-white/30">{index + 1} / {photos.length}</span>
					{#if photo.film_stock}<span>{photo.film_stock}</span>{/if}
					{#if photo.camera}<span>{photo.camera}</span>{/if}
					{#if photo.lens}<span>{photo.lens}</span>{/if}
					{#if photo.location}<span>{photo.location}</span>{/if}
					<button class="ml-auto text-white/40 hover:text-white transition-colors" onclick={copyLink}>{#if copied}Copied!{:else}<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4 inline -mt-0.5"><path d="M12.232 4.232a2.5 2.5 0 013.536 3.536l-1.225 1.224a.75.75 0 001.061 1.06l1.224-1.224a4 4 0 00-5.656-5.656l-3 3a4 4 0 00.225 5.865.75.75 0 00.977-1.138 2.5 2.5 0 01-.142-3.667l3-3z" /><path d="M11.603 7.963a.75.75 0 00-.977 1.138 2.5 2.5 0 01.142 3.667l-3 3a2.5 2.5 0 01-3.536-3.536l1.225-1.224a.75.75 0 00-1.061-1.06l-1.224 1.224a4 4 0 005.656 5.656l3-3a4 4 0 00-.225-5.865z" /></svg>{/if}</button>
					<button class="text-white/40 hover:text-white transition-colors" onclick={() => showMeta = false}>Hide info</button>
				</div>
			</div>
		{:else}
			<div class="flex items-center justify-between p-4">
				<span class="text-white/30 text-sm">{index + 1} / {photos.length}</span>
				<div class="flex items-center gap-4">
					<button class="text-white/40 hover:text-white text-sm transition-colors" onclick={copyLink}>{#if copied}Copied!{:else}<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-4 h-4 inline -mt-0.5"><path d="M12.232 4.232a2.5 2.5 0 013.536 3.536l-1.225 1.224a.75.75 0 001.061 1.06l1.224-1.224a4 4 0 00-5.656-5.656l-3 3a4 4 0 00.225 5.865.75.75 0 00.977-1.138 2.5 2.5 0 01-.142-3.667l3-3z" /><path d="M11.603 7.963a.75.75 0 00-.977 1.138 2.5 2.5 0 01.142 3.667l-3 3a2.5 2.5 0 01-3.536-3.536l1.225-1.224a.75.75 0 00-1.061-1.06l-1.224 1.224a4 4 0 005.656 5.656l3-3a4 4 0 00-.225-5.865z" /></svg>{/if}</button>
					<button class="text-white/40 hover:text-white text-sm transition-colors" onclick={() => showMeta = true}>Info</button>
				</div>
			</div>
		{/if}
	</div>
</div>
