<script lang="ts">
	import { api } from '$lib/api';
	import type { SiteSettings, SocialLink } from '$lib/types';
	import { toasts } from '$lib/stores/toast';
	import SocialLinksEditor from '$lib/components/admin/SocialLinksEditor.svelte';

	let loading = $state(true);
	let saving = $state(false);
	let saved = $state(false);

	let siteTitle = $state('');
	let siteDescription = $state('');
	let photographerName = $state('');
	let aboutText = $state('');
	let accentColor = $state('#f5f5f5');
	let socialLinks = $state<SocialLink[]>([]);

	$effect(() => {
		loadSettings();
	});

	async function loadSettings() {
		try {
			const settings = await api.getSiteSettings();
			siteTitle = settings.site_title || '';
			siteDescription = settings.site_description || '';
			photographerName = settings.photographer_name || '';
			aboutText = settings.about_text || '';
			accentColor = settings.accent_color || '#f5f5f5';
			socialLinks = settings.social_links || [];
		} catch (e) {
			console.error('Failed to load settings:', e);
			toasts.error('Failed to load settings');
		} finally {
			loading = false;
		}
	}

	async function handleSave() {
		saving = true;
		saved = false;
		try {
			await api.updateSiteSettings({
				site_title: siteTitle,
				site_description: siteDescription,
				photographer_name: photographerName,
				about_text: aboutText,
				accent_color: accentColor,
				social_links: socialLinks.filter(l => l.platform && l.url)
			});
			saved = true;
			setTimeout(() => saved = false, 3000);
		} catch (e) {
			console.error('Failed to save settings:', e);
			toasts.error('Failed to save settings');
		} finally {
			saving = false;
		}
	}
</script>

<div class="max-w-2xl">
	<h1 class="text-2xl font-medium mb-6">Site Settings</h1>

	{#if loading}
		<p class="text-text-muted">Loading...</p>
	{:else}
		<div class="space-y-6">
			<div class="bg-surface border border-border rounded-lg p-6 space-y-4">
				<h2 class="font-medium">General</h2>
				<div>
					<label for="site-title" class="block text-sm text-text-muted mb-1">Gallery Name</label>
					<input id="site-title" bind:value={siteTitle} class="w-full px-3 py-2 bg-bg border border-border rounded-md focus:outline-none focus:border-accent" />
				</div>
				<div>
					<label for="site-desc" class="block text-sm text-text-muted mb-1">Description</label>
					<input id="site-desc" bind:value={siteDescription} class="w-full px-3 py-2 bg-bg border border-border rounded-md focus:outline-none focus:border-accent" />
				</div>
				<div>
					<label for="photographer" class="block text-sm text-text-muted mb-1">Photographer Name</label>
					<input id="photographer" bind:value={photographerName} class="w-full px-3 py-2 bg-bg border border-border rounded-md focus:outline-none focus:border-accent" />
				</div>
			</div>

			<div class="bg-surface border border-border rounded-lg p-6 space-y-4">
				<h2 class="font-medium">About</h2>
				<textarea
					bind:value={aboutText}
					rows="6"
					placeholder="Tell visitors about yourself..."
					class="w-full px-3 py-2 bg-bg border border-border rounded-md focus:outline-none focus:border-accent resize-none"
				></textarea>
			</div>

			<div class="bg-surface border border-border rounded-lg p-6 space-y-4">
				<h2 class="font-medium">Social Links</h2>
				<SocialLinksEditor bind:links={socialLinks} />
			</div>

			<div class="bg-surface border border-border rounded-lg p-6 space-y-4">
				<h2 class="font-medium">Theme</h2>
				<div class="flex items-center gap-3">
					<label for="accent" class="text-sm text-text-muted">Accent Color</label>
					<input id="accent" type="color" bind:value={accentColor} class="w-10 h-10 rounded cursor-pointer bg-transparent border-none" />
					<span class="text-sm text-text-muted font-mono">{accentColor}</span>
				</div>
			</div>

			<button
				onclick={handleSave}
				disabled={saving}
				class="px-6 py-2 bg-text text-bg rounded-md font-medium hover:bg-text/90 transition-colors disabled:opacity-50"
			>
				{saving ? 'Saving...' : saved ? 'Saved!' : 'Save Settings'}
			</button>
		</div>
	{/if}
</div>
