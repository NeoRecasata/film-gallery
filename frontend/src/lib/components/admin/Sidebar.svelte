<script lang="ts">
	import { page } from '$app/state';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let { open = false, onclose }: { open?: boolean; onclose?: () => void } = $props();

	const links = [
		{ href: '/admin', label: 'Dashboard', icon: '~' },
		{ href: '/admin/rolls', label: 'Rolls', icon: '#' },
		{ href: '/admin/collections', label: 'Collections', icon: '@' },
		{ href: '/admin/settings', label: 'Settings', icon: '*' },
		{ href: '/admin/account', label: 'Account', icon: '>' }
	];

	async function handleLogout() {
		try {
			await api.logout();
		} catch { /* ignore */ }
		goto('/admin/login');
	}

	function handleNavClick() {
		onclose?.();
	}
</script>

<!-- Mobile backdrop -->
{#if open}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div
		class="fixed inset-0 bg-black/50 z-40 lg:hidden"
		onclick={onclose}
	></div>
{/if}

<aside class="
	bg-surface border-r border-border h-full flex flex-col z-50
	fixed top-0 left-0 w-56 transition-transform duration-200
	{open ? 'translate-x-0' : '-translate-x-full'}
	lg:translate-x-0 lg:static lg:transition-none
">
	<div class="p-4 border-b border-border flex items-center justify-between">
		<a href="/admin" class="text-sm font-semibold tracking-wider uppercase" onclick={handleNavClick}>Admin</a>
		<button onclick={onclose} class="lg:hidden text-text-muted hover:text-text text-lg">&times;</button>
	</div>

	<nav class="flex-1 p-2 space-y-1 overflow-y-auto">
		{#each links as link}
			<a
				href={link.href}
				onclick={handleNavClick}
				class="flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors
					{(link.href === '/admin' ? page.url.pathname === '/admin' : page.url.pathname.startsWith(link.href)) ? 'bg-surface-hover text-text' : 'text-text-muted hover:text-text hover:bg-surface-hover'}"
			>
				<span class="font-mono text-xs w-4 text-center">{link.icon}</span>
				{link.label}
			</a>
		{/each}
	</nav>

	<div class="p-4 border-t border-border">
		<button
			onclick={handleLogout}
			class="w-full text-left text-sm text-text-muted hover:text-text transition-colors"
		>
			Log out
		</button>
	</div>

	<div class="p-4 border-t border-border">
		<a href="/" target="_blank" class="text-xs text-text-muted hover:text-text transition-colors">
			View gallery &rarr;
		</a>
	</div>
</aside>
