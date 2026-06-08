<script lang="ts">
	import type { Snippet } from 'svelte';
	import { setContext } from 'svelte';
	import { SELECT_CONTEXT_KEY, type SelectContext } from './select-context';

	let {
		value = '',
		onValueChange,
		children
	}: {
		value?: string;
		onValueChange?: (value: string) => void;
		children?: Snippet;
	} = $props();

	const labels = new Map<string, string>();

	setContext<SelectContext>(SELECT_CONTEXT_KEY, {
		get value() {
			return value;
		},
		setValue: (next: string) => onValueChange?.(next),
		registerLabel: (itemValue: string, label: string) => {
			labels.set(itemValue, label);
		},
		getLabel: (itemValue: string) => labels.get(itemValue) ?? ''
	});
</script>

<div class="relative">
	{@render children?.()}
</div>
