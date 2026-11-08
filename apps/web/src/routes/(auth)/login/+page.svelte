<script lang="ts">
	import { enhance } from '$app/forms';
	import type { ActionData, PageData } from './$types';
	import { Alert, Button, Input, Label } from '@stery/kit/ui';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let loading = $state(false);
</script>

<div class="min-h-screen flex items-center justify-center bg-background px-4">
	<div class="max-w-md w-full space-y-8">
		<!-- Header -->
		<div class="text-center">
			<h2 class="mt-6 text-3xl font-bold text-foreground">Sign in to HRIS</h2>
			<p class="mt-2 text-sm text-muted-foreground">Multi-tenant HR Management System</p>
		</div>

		<!-- Error Messages -->
		{#if data.error}
			<Alert tone="destructive">
				{#if data.error === 'invalid_session'}
					Your session is invalid. Please log in again.
				{:else}
					{data.error}
				{/if}
			</Alert>
		{/if}

		{#if form?.error}
			<Alert tone="destructive">{form.error}</Alert>
		{/if}

		<!-- Login Form -->
		<form
			method="POST"
			action="?/login"
			use:enhance={() => {
				loading = true;
				return async ({ update }) => {
					await update();
					loading = false;
				};
			}}
			class="space-y-6"
		>
			<input type="hidden" name="redirect" value={data.redirect ?? ''} />
			<!-- Organization Field -->
			<div>
				<Label for="tenant_slug">Organization</Label>
				<Input
					id="tenant_slug"
					type="text"
					name="tenant_slug"
					required
					autocomplete="organization"
					placeholder="your-company"
					disabled={loading}
					class="mt-1"
				/>
			</div>

			<!-- Email Field -->
			<div>
				<Label for="email">Email address</Label>
				<Input
					id="email"
					type="email"
					name="email"
					required
					placeholder="you@example.com"
					disabled={loading}
					class="mt-1"
				/>
			</div>

			<!-- Password Field -->
			<div>
				<Label for="password">Password</Label>
				<Input
					id="password"
					type="password"
					name="password"
					required
					placeholder="••••••••"
					disabled={loading}
					class="mt-1"
				/>
			</div>

			<!-- Submit Button -->
			<Button type="submit" class="w-full" {loading}>
				{loading ? 'Signing in...' : 'Sign in'}
			</Button>
		</form>

	</div>
</div>
