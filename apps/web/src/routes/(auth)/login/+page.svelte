<script lang="ts">
	import { enhance } from '$app/forms';
	import type { PageData } from './$types';
	import { Button } from '$lib/components/ui/button';
	import { AlertCircle } from 'lucide-svelte';

	export let data: PageData;
	export let form: any;

	let loading = false;

	$: if (form?.success) {
		loading = false;
	}
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-50 px-4">
	<div class="max-w-md w-full space-y-8">
		<!-- Header -->
		<div class="text-center">
			<h2 class="mt-6 text-3xl font-bold text-gray-900">Sign in to HRIS</h2>
			<p class="mt-2 text-sm text-gray-600">Multi-tenant HR Management System</p>
		</div>

		<!-- Error Messages -->
		{#if data.error}
			<div class="rounded-md bg-red-50 p-4 flex gap-3">
				<AlertCircle class="w-5 h-5 text-red-600 flex-shrink-0" />
				<div class="text-sm text-red-700">
					{#if data.error === 'invalid_session'}
						Your session is invalid. Please log in again.
					{:else}
						{data.error}
					{/if}
				</div>
			</div>
		{/if}

		{#if form?.error}
			<div class="rounded-md bg-red-50 p-4 flex gap-3">
				<AlertCircle class="w-5 h-5 text-red-600 flex-shrink-0" />
				<div class="text-sm text-red-700">{form.error}</div>
			</div>
		{/if}

		<!-- Login Form -->
		<form method="POST" action="?/login" use:enhance={() => {loading = true}} class="space-y-6">
			<!-- Email Field -->
			<div>
				<label for="email" class="block text-sm font-medium text-gray-700">Email address</label>
				<input
					id="email"
					type="email"
					name="email"
					required
					placeholder="you@example.com"
					disabled={loading}
					class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
				/>
			</div>

			<!-- Password Field -->
			<div>
				<label for="password" class="block text-sm font-medium text-gray-700">Password</label>
				<input
					id="password"
					type="password"
					name="password"
					required
					placeholder="••••••••"
					disabled={loading}
					class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
				/>
			</div>

			<!-- Submit Button -->
			<Button type="submit" class="w-full" disabled={loading}>
				{#if loading}
					<span class="inline-block animate-spin mr-2">⟳</span>
					Signing in...
				{:else}
					Sign in
				{/if}
			</Button>
		</form>

		<!-- Demo Credentials -->
		<div class="rounded-md bg-blue-50 p-4">
			<p class="text-xs font-medium text-blue-700">Demo Credentials (Development):</p>
			<ul class="mt-2 text-xs text-blue-600 space-y-1">
				<li><strong>Employee:</strong> employee@example.com / password</li>
				<li><strong>Manager:</strong> manager@example.com / password</li>
				<li><strong>HR Admin:</strong> admin@example.com / password</li>
			</ul>
		</div>
	</div>
</div>

<style>
	:global(body) {
		background-color: rgb(249 250 251);
	}
</style>
