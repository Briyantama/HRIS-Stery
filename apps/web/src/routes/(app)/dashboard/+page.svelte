<script lang="ts">
	import { user } from '$lib/stores/auth';
	import { useLeaveBalance } from '$lib/queries/leave';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';

	// Get current user ID from store
	let userId: string | null = null;
	user.subscribe((u) => {
		userId = u?.user_id ?? null;
	});

	// Fetch leave balance
	$: leaveBalance = useLeaveBalance(userId);

	// Format date for display
	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	// Get badge color based on remaining days
	function getRemainingDaysColor(remaining: number, entitled: number): string {
		const percentage = (remaining / entitled) * 100;
		if (percentage > 50) return 'bg-green-100 text-green-800';
		if (percentage > 25) return 'bg-yellow-100 text-yellow-800';
		return 'bg-red-100 text-red-800';
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div>
		<h1 class="text-3xl font-bold tracking-tight">Leave Balance</h1>
		<p class="mt-2 text-gray-600">Your leave entitlements and usage for {new Date().getFullYear()}</p>
	</div>

	<!-- Loading State -->
	{#if $leaveBalance.isLoading}
		<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
			{#each Array(4) as _}
				<Card>
					<CardHeader class="pb-2">
						<Skeleton class="h-4 w-24" />
					</CardHeader>
					<CardContent>
						<Skeleton class="h-8 w-12" />
						<Skeleton class="mt-2 h-3 w-32" />
					</CardContent>
				</Card>
			{/each}
		</div>
	{/if}

	<!-- Error State -->
	{#if $leaveBalance.isError}
		<div class="rounded-lg border border-red-200 bg-red-50 p-4">
			<p class="text-sm font-medium text-red-800">Unable to load leave balance</p>
			<p class="mt-1 text-sm text-red-600">
				{$leaveBalance.error instanceof Error ? $leaveBalance.error.message : 'Unknown error'}
			</p>
		</div>
	{/if}

	<!-- Success State -->
	{#if $leaveBalance.data}
		{@const balance = $leaveBalance.data}

		<!-- Year Info -->
		<div class="text-sm text-gray-600">
			Viewing leave balance for year: <span class="font-semibold">{balance.year}</span>
		</div>

		<!-- Summary Cards Grid -->
		{#if balance.balances && balance.balances.length > 0}
			<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
				{#each balance.balances as typeBalance (typeBalance.leave_type_id)}
					<Card>
						<CardHeader class="pb-3">
							<CardTitle class="text-base">{typeBalance.leave_type_name}</CardTitle>
							<CardDescription>{typeBalance.leave_type_code}</CardDescription>
						</CardHeader>
						<CardContent class="space-y-4">
							<!-- Entitled Days -->
							<div>
								<p class="text-xs font-medium text-gray-500 uppercase">Entitled</p>
								<p class="mt-1 text-2xl font-bold">{typeBalance.entitled_days}</p>
							</div>

							<!-- Usage Breakdown -->
							<div class="space-y-2 border-t pt-3">
								<div class="flex justify-between text-sm">
									<span class="text-gray-600">Used</span>
									<span class="font-medium">{typeBalance.used_days}</span>
								</div>
								<div class="flex justify-between text-sm">
									<span class="text-gray-600">Pending</span>
									<span class="font-medium text-yellow-600">{typeBalance.pending_days}</span>
								</div>
							</div>

							<!-- Remaining Days Badge -->
							<div
								class={`rounded-lg px-3 py-2 text-center font-semibold ${getRemainingDaysColor(
									typeBalance.remaining_days,
									typeBalance.entitled_days
								)}`}
							>
								<p class="text-xs font-medium opacity-75">Remaining</p>
								<p class="text-lg">{typeBalance.remaining_days}</p>
							</div>

							<!-- Progress Bar -->
							<div class="space-y-1">
								<p class="text-xs text-gray-500">Usage</p>
								<div class="h-2 overflow-hidden rounded-full bg-gray-200">
									<div
										class="h-full bg-blue-500 transition-all"
										style="width: {((typeBalance.used_days + typeBalance.pending_days) /
											typeBalance.entitled_days) *
											100}%"
									/>
								</div>
								<p class="text-right text-xs text-gray-500">
									{typeBalance.used_days + typeBalance.pending_days} / {typeBalance.entitled_days}
								</p>
							</div>
						</CardContent>
					</Card>
				{/each}
			</div>
		{:else}
			<div class="rounded-lg border border-gray-200 bg-gray-50 p-8 text-center">
				<p class="text-gray-600">No leave types configured for your tenant</p>
			</div>
		{/if}

		<!-- Legend -->
		<div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
			<p class="text-sm font-medium text-gray-700">Legend</p>
			<div class="mt-3 space-y-2 text-sm text-gray-600">
				<div class="flex items-center gap-2">
					<span class="inline-block h-3 w-3 rounded bg-blue-500" />
					<span><strong>Used Days:</strong> Already taken leave</span>
				</div>
				<div class="flex items-center gap-2">
					<span class="inline-block h-3 w-3 rounded bg-yellow-400" />
					<span><strong>Pending:</strong> Approved requests awaiting start date</span>
				</div>
				<div class="flex items-center gap-2">
					<span class="inline-block h-3 w-3 rounded" style="background-color: #f0f0f0; border: 1px solid #d0d0d0;" />
					<span><strong>Remaining:</strong> Available days to use</span>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	:global(body) {
		/* Ensure consistent spacing */
	}
</style>
