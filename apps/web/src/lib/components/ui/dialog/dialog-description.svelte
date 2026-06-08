<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children }: { children?: Snippet } = $props();
</script>

<p class="text-sm text-gray-600">
	{@render children?.()}
</p>
