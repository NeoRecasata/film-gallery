<script lang="ts">
	import type { Snippet } from 'svelte';

	let { open, onclose, children }: { open: boolean; onclose: () => void; children: Snippet } = $props();

	function handleBackdrop(e: MouseEvent) {
		if (e.target === e.currentTarget) onclose();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onclose();
	}
</script>

{#if open}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm"
		onclick={handleBackdrop}
		onkeydown={handleKeydown}
	>
		<div class="absolute bottom-0 left-0 right-0 max-h-[70vh] bg-surface border-t border-border rounded-t-2xl overflow-y-auto animate-slide-up">
			<div class="sticky top-0 bg-surface pt-3 pb-2 flex justify-center">
				<div class="w-10 h-1 rounded-full bg-border"></div>
			</div>
			<div class="px-4 pb-6">
				{@render children()}
			</div>
		</div>
	</div>
{/if}

<style>
	@keyframes slide-up {
		from { transform: translateY(100%); }
		to { transform: translateY(0); }
	}
	.animate-slide-up {
		animation: slide-up 0.25s ease-out;
	}
</style>
