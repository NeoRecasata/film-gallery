<script lang="ts">
	import { toasts } from '$lib/stores/toast';
	import { fly } from 'svelte/transition';

	let items = $state<{ id: number; message: string; type: 'success' | 'error' }[]>([]);

	toasts.subscribe((value) => {
		items = value;
	});
</script>

{#if items.length > 0}
	<div class="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none">
		{#each items as toast (toast.id)}
			<div
				transition:fly={{ x: 100, duration: 250 }}
				class="pointer-events-auto px-4 py-3 rounded-lg shadow-lg text-sm font-medium flex items-center gap-3 min-w-[280px] max-w-[400px]
					{toast.type === 'success' ? 'bg-success/15 text-success border border-success/30' : 'bg-error/15 text-error border border-error/30'}"
			>
				<span class="flex-shrink-0 text-base">
					{toast.type === 'success' ? '\u2713' : '\u2717'}
				</span>
				<span class="flex-1">{toast.message}</span>
				<button
					onclick={() => toasts.dismiss(toast.id)}
					class="flex-shrink-0 opacity-60 hover:opacity-100 transition-opacity text-base leading-none"
				>
					&times;
				</button>
			</div>
		{/each}
	</div>
{/if}
