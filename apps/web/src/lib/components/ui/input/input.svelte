<script lang="ts">
	let {
		id = undefined,
		type = 'text',
		value = $bindable(''),
		disabled = false,
		class: className = '',
		onchange
	}: {
		id?: string;
		type?: string;
		value?: string;
		disabled?: boolean;
		class?: string;
		onchange?: (event: Event) => void;
	} = $props();
</script>

<input
	{id}
	{type}
	bind:value
	{disabled}
	class="flex h-10 w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm {className}"
	{onchange}
/>
