<script lang="ts">
	import type { LayoutServerData } from '../../$types';

	let { data }: { data: LayoutServerData } = $props();
	const settings = $derived(data.settings);
</script>

<svelte:head>
	<title>About | {settings.site_title}</title>
</svelte:head>

<div class="max-w-2xl mx-auto px-4 py-16">
	<h1 class="text-3xl font-medium mb-6">
		{settings.photographer_name || 'About'}
	</h1>

	{#if settings.about_text}
		<div class="prose prose-invert max-w-none text-text-muted leading-relaxed whitespace-pre-line">
			{settings.about_text}
		</div>
	{/if}

	{#if settings.social_links.length > 0}
		<div class="mt-10 flex flex-wrap gap-4">
			{#each settings.social_links as link}
				<a
					href={link.url}
					target="_blank"
					rel="noopener noreferrer"
					class="px-4 py-2 bg-surface rounded-lg text-sm text-text-muted hover:text-text hover:bg-surface-hover transition-colors capitalize"
				>
					{link.platform}
				</a>
			{/each}
		</div>
	{/if}
</div>
