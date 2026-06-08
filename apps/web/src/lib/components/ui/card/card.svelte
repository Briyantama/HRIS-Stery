<script lang="ts">
	import type { Snippet } from 'svelte';

	let { class: className = '', children }: { class?: string; children?: Snippet } = $props();
</script>

<div class="rounded-lg border bg-white shadow-sm {className}">
	{@render children?.()}
</div>
