<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		open = $bindable(false),
		children
	}: {
		open?: boolean;
		children?: Snippet;
	} = $props();
</script>

{#if open}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
		{@render children?.()}
	</div>
{/if}
