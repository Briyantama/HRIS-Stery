<script lang="ts">
	import { page } from '$app/stores';
	import { Button } from '@stery/kit/ui';
	import { CircleAlert } from '@lucide/svelte';
</script>

<div class="min-h-screen flex items-center justify-center bg-background px-4">
	<div class="max-w-md w-full text-center">
		<div class="flex justify-center mb-6">
			<div class="p-3 bg-destructive/10 rounded-full">
				<CircleAlert class="w-8 h-8 text-destructive" />
			</div>
		</div>

		<h1 class="text-3xl font-bold text-foreground mb-2">Access Denied</h1>

		<p class="text-muted-foreground mb-6">
			You don't have permission to access this resource. Only managers and HR administrators can access
			this page.
		</p>

		{#if $page.url.searchParams.get('requested')}
			<p class="text-sm text-muted-foreground mb-6">
				Requested page: <code class="bg-muted px-2 py-1 rounded">{$page.url.searchParams.get('requested')}</code>
			</p>
		{/if}

		<div class="space-y-3">
			<Button href="/dashboard" class="w-full" variant="primary">
				Go to Dashboard
			</Button>
			<Button href="/" class="w-full" variant="outline">
				Go Home
			</Button>
		</div>

		<p class="text-xs text-muted-foreground mt-6">
			If you believe this is an error, please contact your administrator.
		</p>
	</div>
</div>
