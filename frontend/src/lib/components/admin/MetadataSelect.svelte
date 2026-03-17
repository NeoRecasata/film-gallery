<script lang="ts">
	let { value = $bindable(''), options = [], placeholder = '' }: { value: string; options: string[]; placeholder?: string } = $props();

	let customMode = $state(false);

	// If current value isn't in options, show text input
	const isCustom = $derived(customMode || (value !== '' && !options.includes(value)));

	function handleSelect(e: Event) {
		const selected = (e.target as HTMLSelectElement).value;
		if (selected === '__new__') {
			customMode = true;
			value = '';
		} else {
			customMode = false;
			value = selected;
		}
	}

	function switchToSelect() {
		customMode = false;
		if (!options.includes(value)) {
			value = '';
		}
	}
</script>

{#if isCustom}
	<div class="flex gap-1">
		<input
			bind:value
			{placeholder}
			class="flex-1 min-w-0 px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent placeholder:text-text-muted/40"
		/>
		{#if options.length > 0}
			<button
				type="button"
				onclick={switchToSelect}
				class="px-2 py-2 bg-bg border border-border rounded-md text-text-muted hover:text-text text-xs transition-colors flex-shrink-0"
				title="Choose from list"
			>
				<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="w-3.5 h-3.5">
					<path fill-rule="evenodd" d="M4.22 6.22a.75.75 0 0 1 1.06 0L8 8.94l2.72-2.72a.75.75 0 1 1 1.06 1.06l-3.25 3.25a.75.75 0 0 1-1.06 0L4.22 7.28a.75.75 0 0 1 0-1.06Z" clip-rule="evenodd" />
				</svg>
			</button>
		{/if}
	</div>
{:else}
	<select
		{value}
		onchange={handleSelect}
		class="w-full px-3 py-2 bg-bg border border-border rounded-md text-sm focus:outline-none focus:border-accent {value === '' ? 'text-text-muted/40' : 'text-text'}"
	>
		<option value="" disabled>{placeholder}</option>
		{#each options as opt}
			<option value={opt}>{opt}</option>
		{/each}
		<option value="__new__">+ New...</option>
	</select>
{/if}
