<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children }: { children?: Snippet } = $props();
</script>

<div class="mt-1 max-h-60 w-full overflow-auto rounded-md border border-gray-200 bg-white shadow-md">
	{@render children?.()}
</div>
