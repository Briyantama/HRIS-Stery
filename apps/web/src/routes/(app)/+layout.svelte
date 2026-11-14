<script lang="ts">
	import type { Snippet } from 'svelte';
	import { page } from '$app/state';
	import { isManager } from '$lib/stores/auth';
	import { AppShell, SidebarNav } from '@stery/kit/ui';

	let { children }: { children: Snippet } = $props();

	const items = $derived(
		[
			{ label: 'Dashboard', href: '/dashboard' },
			{ label: 'Request Leave', href: '/leaves/apply' },
			...($isManager ? [{ label: 'Leave Approval', href: '/admin/leaves' }] : [])
		].map((item) => ({ ...item, active: page.url.pathname === item.href }))
	);
</script>

<AppShell>
	{#snippet sidebar()}
		<SidebarNav aria-label="Main" {items} />
	{/snippet}
	{#snippet header()}
		<span class="font-semibold text-foreground">HRIS</span>
	{/snippet}
	{@render children()}
</AppShell>
