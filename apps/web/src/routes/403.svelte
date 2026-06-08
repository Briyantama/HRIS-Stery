<script lang="ts">
	import { page } from '$app/stores';
	import { Button } from '$lib/components/ui/button';
	import { AlertCircle } from 'lucide-svelte';
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-50 px-4">
	<div class="max-w-md w-full text-center">
		<div class="flex justify-center mb-6">
			<div class="p-3 bg-red-100 rounded-full">
				<AlertCircle class="w-8 h-8 text-red-600" />
			</div>
		</div>

		<h1 class="text-3xl font-bold text-gray-900 mb-2">Access Denied</h1>

		<p class="text-gray-600 mb-6">
			You don't have permission to access this resource. Only managers and HR administrators can access
			this page.
		</p>

		{#if $page.url.searchParams.get('requested')}
			<p class="text-sm text-gray-500 mb-6">
				Requested page: <code class="bg-gray-100 px-2 py-1 rounded">{$page.url.searchParams.get('requested')}</code>
			</p>
		{/if}

		<div class="space-y-3">
			<Button href="/dashboard" class="w-full" variant="default">
				Go to Dashboard
			</Button>
			<Button href="/" class="w-full" variant="outline">
				Go Home
			</Button>
		</div>

		<p class="text-xs text-gray-500 mt-6">
			If you believe this is an error, please contact your administrator.
		</p>
	</div>
</div>

<style>
	:global(body) {
		background-color: rgb(249 250 251);
	}
</style>
