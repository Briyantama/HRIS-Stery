<script lang="ts">
	import { user } from '$lib/stores/auth';
	import { useLeaveBalance, useLeaveTypes } from '$lib/queries/leave';
	import { Alert, Card, CardContent, CardDescription, CardHeader, CardTitle, Skeleton } from '@stery/kit/ui';

	// Get current user ID from store
	let userId: string | null = null;
	user.subscribe((u) => {
		userId = u?.user_id ?? null;
	});

	// Fetch leave balance
	$: leaveBalance = useLeaveBalance(userId);

	// The balance only carries leave_type_id; names and codes come from the tenant's leave types.
	const leaveTypes = useLeaveTypes();
	$: typeById = new Map(($leaveTypes.data ?? []).map((t) => [t.id, t]));

	// Get badge color based on remaining days
	function getRemainingDaysColor(remaining: number, entitled: number): string {
		const percentage = (remaining / entitled) * 100;
		if (percentage > 50) return 'bg-success/15 text-success';
		if (percentage > 25) return 'bg-warning/15 text-warning';
		return 'bg-destructive/10 text-destructive';
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div>
		<h1 class="text-3xl font-bold tracking-tight">Leave Balance</h1>
		<p class="mt-2 text-muted-foreground">Your leave entitlements and usage for {new Date().getFullYear()}</p>
	</div>

	<!-- Loading State -->
	{#if $leaveBalance.isLoading}
		<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
			{#each [0, 1, 2, 3] as i (i)}
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
		<Alert tone="destructive" title="Unable to load leave balance">
			<p class="text-muted-foreground">
				{$leaveBalance.error instanceof Error ? $leaveBalance.error.message : 'Unknown error'}
			</p>
		</Alert>
	{/if}

	<!-- Success State -->
	{#if $leaveBalance.data}
		{@const balance = $leaveBalance.data}

		<!-- Year Info -->
		<div class="text-sm text-muted-foreground">
			Viewing leave balance for year: <span class="font-semibold">{balance.year}</span>
		</div>

		<!-- Summary Cards Grid -->
		{#if balance.by_type.length > 0}
			<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
				{#each balance.by_type as typeBalance (typeBalance.leave_type_id)}
					<Card>
						<CardHeader class="pb-3">
							<CardTitle class="text-base">{typeById.get(typeBalance.leave_type_id)?.name ?? 'Leave'}</CardTitle>
							<CardDescription>{typeById.get(typeBalance.leave_type_id)?.code ?? ''}</CardDescription>
						</CardHeader>
						<CardContent class="space-y-4">
							<!-- Entitled Days -->
							<div>
								<p class="text-xs font-medium text-muted-foreground uppercase">Entitled</p>
								<p class="mt-1 text-2xl font-bold">{typeBalance.entitled_days}</p>
							</div>

							<!-- Usage Breakdown -->
							<div class="space-y-2 border-t pt-3">
								<div class="flex justify-between text-sm">
									<span class="text-muted-foreground">Used</span>
									<span class="font-medium">{typeBalance.used_days}</span>
								</div>
								<div class="flex justify-between text-sm">
									<span class="text-muted-foreground">Pending</span>
									<span class="font-medium text-warning">{typeBalance.pending_days}</span>
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
								<p class="text-xs text-muted-foreground">Usage</p>
								<div class="h-2 overflow-hidden rounded-full bg-muted">
									<div
										class="h-full bg-primary transition-all"
										style="width: {((typeBalance.used_days + typeBalance.pending_days) /
											typeBalance.entitled_days) *
											100}%"
									></div>
								</div>
								<p class="text-right text-xs text-muted-foreground">
									{typeBalance.used_days + typeBalance.pending_days} / {typeBalance.entitled_days}
								</p>
							</div>
						</CardContent>
					</Card>
				{/each}
			</div>
		{:else}
			<div class="rounded-lg border border-border bg-muted p-8 text-center">
				<p class="text-muted-foreground">No leave types configured for your tenant</p>
			</div>
		{/if}

		<!-- Legend -->
		<div class="rounded-lg border border-border bg-muted p-4">
			<p class="text-sm font-medium text-foreground">Legend</p>
			<div class="mt-3 space-y-2 text-sm text-muted-foreground">
				<div class="flex items-center gap-2">
					<span class="inline-block h-3 w-3 rounded bg-primary"></span>
					<span><strong>Used Days:</strong> Already taken leave</span>
				</div>
				<div class="flex items-center gap-2">
					<span class="inline-block h-3 w-3 rounded bg-warning"></span>
					<span><strong>Pending:</strong> Approved requests awaiting start date</span>
				</div>
				<div class="flex items-center gap-2">
					<span class="inline-block h-3 w-3 rounded border border-border bg-surface"></span>
					<span><strong>Remaining:</strong> Available days to use</span>
				</div>
			</div>
		</div>
	{/if}
</div>

