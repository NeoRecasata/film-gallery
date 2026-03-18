<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
	import Sidebar from '$lib/components/admin/Sidebar.svelte';
	import Toast from '$lib/components/admin/Toast.svelte';

	let { children }: { children: Snippet } = $props();

	const isAuthPage = $derived(
		page.url.pathname === '/admin/login' || page.url.pathname === '/admin/setup'
	);

	let sidebarOpen = $state(false);
</script>

<svelte:head>
	<title>Admin</title>
</svelte:head>

<Toast />

{#if isAuthPage}
	{@render children()}
{:else}
	<div class="flex h-screen overflow-hidden">
		<Sidebar open={sidebarOpen} onclose={() => sidebarOpen = false} />

		<div class="flex-1 flex flex-col min-w-0">
			<!-- Mobile top bar -->
			<div class="lg:hidden flex items-center gap-3 px-4 py-3 border-b border-border bg-surface flex-shrink-0">
				<button
					onclick={() => sidebarOpen = true}
					class="text-text-muted hover:text-text transition-colors"
					aria-label="Open menu"
				>
					<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5">
						<path fill-rule="evenodd" d="M2 4.75A.75.75 0 012.75 4h14.5a.75.75 0 010 1.5H2.75A.75.75 0 012 4.75zM2 10a.75.75 0 01.75-.75h14.5a.75.75 0 010 1.5H2.75A.75.75 0 012 10zm0 5.25a.75.75 0 01.75-.75h14.5a.75.75 0 010 1.5H2.75a.75.75 0 01-.75-.75z" clip-rule="evenodd" />
					</svg>
				</button>
				<span class="text-sm font-semibold tracking-wider uppercase">Admin</span>
			</div>

			<main class="flex-1 p-4 lg:p-8 overflow-y-auto [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
				{@render children()}
			</main>
		</div>
	</div>
{/if}
