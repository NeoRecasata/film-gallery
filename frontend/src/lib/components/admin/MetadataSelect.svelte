<script lang="ts">
	let { value = $bindable(''), options = [], placeholder = '' }: { value: string; options: string[]; placeholder?: string } = $props();

	let open = $state(false);
	let inputEl: HTMLInputElement | undefined = $state();

	const filtered = $derived(
		value.trim() === ''
			? options
			: options.filter(o => o.toLowerCase().includes(value.toLowerCase()))
	);

	function select(opt: string) {
		value = opt;
		open = false;
	}

	function handleFocus() {
		open = true;
	}

	function handleBlur() {
		// Delay to allow click on dropdown item
		setTimeout(() => { open = false; }, 150);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			open = false;
			inputEl?.blur();
		}
	}
</script>

<div class="relative">
	<input
		bind:this={inputEl}
		bind:value
		{placeholder}
		onfocus={handleFocus}
		onblur={handleBlur}
		onkeydown={handleKeydown}
		autocomplete="off"
		class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40"
	/>

	{#if open && filtered.length > 0}
		<div class="absolute z-50 left-0 right-0 mt-1 max-h-[200px] overflow-y-auto bg-surface border border-border rounded-md shadow-xl [scrollbar-width:thin]">
			{#each filtered as opt}
				<button
					type="button"
					onmousedown={(e) => { e.preventDefault(); select(opt); }}
					class="w-full text-left px-3 py-2 text-sm hover:bg-surface-hover transition-colors
						{opt === value ? 'bg-surface-hover text-text' : 'text-text'}"
				>
					{opt}
				</button>
			{/each}
		</div>
	{/if}
</div>
