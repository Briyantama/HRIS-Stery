<script lang="ts">
	import type { Snippet } from 'svelte';

	type Variant = 'default' | 'outline' | 'destructive';
	type Size = 'sm' | 'default';

	let {
		variant = 'default',
		size = 'default',
		href = undefined,
		type = 'button',
		disabled = false,
		class: className = '',
		children,
		onclick
	}: {
		variant?: Variant;
		size?: Size;
		href?: string;
		type?: 'button' | 'submit';
		disabled?: boolean;
		class?: string;
		children?: Snippet;
		onclick?: (event: MouseEvent) => void;
	} = $props();

	const variants: Record<Variant, string> = {
		default: 'bg-blue-600 text-white hover:bg-blue-700',
		outline: 'border border-gray-300 bg-white text-gray-900 hover:bg-gray-50',
		destructive: 'bg-red-600 text-white hover:bg-red-700'
	};

	const sizes: Record<Size, string> = {
		sm: 'px-3 py-1.5 text-sm',
		default: 'px-4 py-2 text-sm'
	};

	const classes = $derived(
		`inline-flex items-center justify-center rounded-md font-medium transition-colors disabled:opacity-50 ${variants[variant]} ${sizes[size]} ${className}`
	);
</script>

{#if href}
	<a {href} class={classes} class:opacity-50={disabled}>
		{@render children?.()}
	</a>
{:else}
	<button {type} {disabled} class={classes} {onclick}>
		{@render children?.()}
	</button>
{/if}
