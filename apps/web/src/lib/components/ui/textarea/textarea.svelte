<script lang="ts">
	let {
		id = undefined,
		placeholder = '',
		value = $bindable(''),
		disabled = false,
		class: className = ''
	}: {
		id?: string;
		placeholder?: string;
		value?: string;
		disabled?: boolean;
		class?: string;
	} = $props();
</script>

<textarea
	{id}
	{placeholder}
	bind:value
	{disabled}
	class="flex w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm {className}"
></textarea>
