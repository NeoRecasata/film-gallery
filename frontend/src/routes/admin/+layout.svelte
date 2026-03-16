<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
	import Sidebar from '$lib/components/admin/Sidebar.svelte';

	let { children }: { children: Snippet } = $props();

	const isAuthPage = $derived(
		page.url.pathname === '/admin/login' || page.url.pathname === '/admin/setup'
	);
</script>

{#if isAuthPage}
	{@render children()}
{:else}
	<div class="flex min-h-screen">
		<Sidebar />
		<main class="flex-1 p-8 overflow-auto">
			{@render children()}
		</main>
	</div>
{/if}
