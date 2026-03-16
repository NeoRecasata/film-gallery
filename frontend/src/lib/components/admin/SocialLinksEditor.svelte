<script lang="ts">
	import type { SocialLink } from '$lib/types';

	let { links = $bindable([]) }: { links: SocialLink[] } = $props();

	function addLink() {
		links = [...links, { platform: '', url: '' }];
	}

	function removeLink(index: number) {
		links = links.filter((_, i) => i !== index);
	}
</script>

<div class="space-y-2">
	{#each links as link, i}
		<div class="flex gap-2">
			<input
				bind:value={link.platform}
				placeholder="Platform (e.g. instagram)"
				class="flex-1 px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent"
			/>
			<input
				bind:value={link.url}
				placeholder="https://..."
				class="flex-[2] px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent"
			/>
			<button onclick={() => removeLink(i)} class="text-error/60 hover:text-error text-sm px-2">
				Remove
			</button>
		</div>
	{/each}

	<button onclick={addLink} class="text-sm text-text-muted hover:text-text">
		+ Add social link
	</button>
</div>
