<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children }: { children?: Snippet } = $props();
</script>

<h2 class="text-lg font-semibold text-gray-900">
	{@render children?.()}
</h2>
