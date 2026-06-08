<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		class: className = '',
		children
	}: {
		class?: string;
		children?: Snippet;
	} = $props();
</script>

<div class="w-full max-w-lg rounded-lg bg-white p-6 shadow-lg {className}">
	{@render children?.()}
</div>
