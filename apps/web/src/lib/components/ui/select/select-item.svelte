<script lang="ts">
	import type { Snippet } from 'svelte';
	import { getContext } from 'svelte';
	import { SELECT_CONTEXT_KEY, type SelectContext } from './select-context';

	let {
		value,
		children
	}: {
		value: string;
		children?: Snippet;
	} = $props();

	const ctx = getContext<SelectContext>(SELECT_CONTEXT_KEY);
</script>

<button
	type="button"
	class="block w-full px-3 py-2 text-left text-sm hover:bg-gray-100"
	onclick={() => ctx.setValue(value)}
>
	{@render children?.()}
</button>
