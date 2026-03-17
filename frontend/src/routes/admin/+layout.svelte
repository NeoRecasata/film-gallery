<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
	import Sidebar from '$lib/components/admin/Sidebar.svelte';
	import Toast from '$lib/components/admin/Toast.svelte';

	let { children }: { children: Snippet } = $props();

	const isAuthPage = $derived(
		page.url.pathname === '/admin/login' || page.url.pathname === '/admin/setup'
	);
</script>

<svelte:head>
	<title>Admin</title>
</svelte:head>

<Toast />

{#if isAuthPage}
	{@render children()}
{:else}
	<div class="flex h-screen overflow-hidden">
		<Sidebar />
		<main class="flex-1 p-8 overflow-y-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
			{@render children()}
		</main>
	</div>
{/if}
