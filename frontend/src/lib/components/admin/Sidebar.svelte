<script lang="ts">
	import { page } from '$app/state';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	const links = [
		{ href: '/admin', label: 'Dashboard', icon: '~' },
		{ href: '/admin/photos', label: 'Photos', icon: '#' },
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
</script>

<aside class="w-56 bg-surface border-r border-border min-h-screen flex flex-col">
	<div class="p-4 border-b border-border">
		<a href="/admin" class="text-sm font-semibold tracking-wider uppercase">Admin</a>
	</div>

	<nav class="flex-1 p-2 space-y-1">
		{#each links as link}
			<a
				href={link.href}
				class="flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors
					{page.url.pathname === link.href ? 'bg-surface-hover text-text' : 'text-text-muted hover:text-text hover:bg-surface-hover'}"
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
