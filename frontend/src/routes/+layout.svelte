<script lang="ts">
	import '../app.css';
	import type { LayoutServerData } from './$types';

	import type { Snippet } from 'svelte';

	let { data, children }: { data: LayoutServerData; children: Snippet } = $props();
</script>

<svelte:head>
	<title>{data.settings.site_title}</title>
	<meta name="description" content={data.settings.site_description} />
</svelte:head>

{@render children()}
