<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		id = undefined,
		class: className = '',
		children
	}: {
		id?: string;
		class?: string;
		children?: Snippet;
	} = $props();
</script>

<button
	{id}
	type="button"
	class="flex h-10 w-full items-center justify-between rounded-md border border-gray-300 bg-white px-3 py-2 text-sm {className}"
>
	{@render children?.()}
</button>
