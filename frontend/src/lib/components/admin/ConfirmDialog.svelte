<script lang="ts">
	let {
		open = false,
		title = 'Confirm',
		message = 'Are you sure?',
		confirmLabel = 'Delete',
		cancelLabel = 'Cancel',
		variant = 'danger' as 'danger' | 'warning',
		onconfirm,
		oncancel
	}: {
		open: boolean;
		title?: string;
		message?: string;
		confirmLabel?: string;
		cancelLabel?: string;
		variant?: 'danger' | 'warning';
		onconfirm: () => void;
		oncancel: () => void;
	} = $props();

	let dialogEl: HTMLDialogElement | undefined = $state();

	$effect(() => {
		if (!dialogEl) return;
		if (open && !dialogEl.open) {
			dialogEl.showModal();
		} else if (!open && dialogEl.open) {
			dialogEl.close();
		}
	});

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			e.preventDefault();
			oncancel();
		}
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === dialogEl) {
			oncancel();
		}
	}
</script>

{#if open}
	<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
	<dialog
		bind:this={dialogEl}
		onkeydown={handleKeydown}
		onclick={handleBackdropClick}
		class="fixed inset-0 z-[200] bg-transparent backdrop:bg-black/60 p-0 m-auto border-0 outline-none max-w-md w-full"
	>
		<div class="bg-surface border border-border rounded-xl p-6 shadow-2xl">
			<h2 class="text-lg font-semibold text-text mb-2">{title}</h2>
			<p class="text-sm text-text-muted leading-relaxed mb-6">{message}</p>

			<div class="flex items-center justify-end gap-3">
				<button
					onclick={oncancel}
					class="px-4 py-2 rounded-md text-sm text-text-muted border border-border hover:border-text-muted/40 hover:text-text transition-colors"
				>
					{cancelLabel}
				</button>
				<button
					onclick={onconfirm}
					class="px-4 py-2 rounded-md text-sm font-medium transition-colors
						{variant === 'danger' ? 'bg-error hover:bg-error/80 text-white' : 'bg-amber-600 hover:bg-amber-500 text-white'}"
				>
					{confirmLabel}
				</button>
			</div>
		</div>
	</dialog>
{/if}
