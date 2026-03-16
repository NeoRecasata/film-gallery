<script lang="ts">
	import type { LayoutServerData } from '../$types';

	import type { Snippet } from 'svelte';

	let { data, children }: { data: LayoutServerData; children: Snippet } = $props();
	const settings = $derived(data.settings);
</script>

<div class="min-h-screen flex flex-col">
	<nav class="fixed top-0 left-0 right-0 z-50 bg-bg/80 backdrop-blur-md border-b border-border">
		<div class="max-w-7xl mx-auto px-4 h-14 flex items-center justify-between">
			<a href="/" class="text-sm font-semibold tracking-widest uppercase text-text hover:text-white transition-colors">
				{settings.photographer_name || settings.site_title}
			</a>
			<div class="flex items-center gap-6 text-sm text-text-muted">
				<a href="/collections" class="hover:text-text transition-colors">Collections</a>
				<a href="/about" class="hover:text-text transition-colors">About</a>
			</div>
		</div>
	</nav>

	<main class="flex-1 pt-14">
		{@render children()}
	</main>

	<footer class="border-t border-border py-8 text-center text-sm text-text-muted">
		<p>&copy; {new Date().getFullYear()} {settings.photographer_name || settings.site_title}</p>
	</footer>
</div>
