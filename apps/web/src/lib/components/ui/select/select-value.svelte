<script lang="ts">
	import { getContext } from 'svelte';
	import { SELECT_CONTEXT_KEY, type SelectContext } from './select-context';

	let { placeholder = '' }: { placeholder?: string } = $props();

	const ctx = getContext<SelectContext>(SELECT_CONTEXT_KEY);
</script>

<span class="text-gray-900">
	{ctx.value ? ctx.getLabel(ctx.value) || ctx.value : placeholder}
</span>
