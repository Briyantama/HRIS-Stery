<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		for: htmlFor = undefined,
		class: className = '',
		children
	}: {
		for?: string;
		class?: string;
		children?: Snippet;
	} = $props();
</script>

<label for={htmlFor} class="text-sm font-medium text-gray-900 {className}">
	{@render children?.()}
</label>
