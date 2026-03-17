<script lang="ts">
	let {
		hasMore = false,
		loading = false,
		onloadmore
	}: {
		hasMore: boolean;
		loading: boolean;
		onloadmore: () => void;
	} = $props();

	let sentinel: HTMLDivElement | undefined = $state();

	$effect(() => {
		if (!sentinel || !hasMore) return;

		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0].isIntersecting && !loading && hasMore) {
					onloadmore();
				}
			},
			{ rootMargin: '200px' }
		);

		observer.observe(sentinel);
		return () => observer.disconnect();
	});
</script>

<div bind:this={sentinel} class="w-full py-8 flex justify-center">
	{#if loading}
		<div class="w-6 h-6 border-2 border-text-muted border-t-transparent rounded-full animate-spin"></div>
	{/if}
</div>
