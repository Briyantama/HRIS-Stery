<script lang="ts">
	import { derived } from 'svelte/store';
	import { user } from '$lib/stores/auth';
	import { useLeaveBalance, useLeaveTypes } from '$lib/queries/leave';
	import { Alert, Card, CardContent, CardDescription, CardHeader, CardTitle, Skeleton } from '@stery/kit/ui';

	const leaveBalance = useLeaveBalance(derived(user, (u) => u?.user_id ?? null));

	// The balance only carries leave_type_id; names and codes come from the tenant's leave types.
	const leaveTypes = useLeaveTypes();
	$: typeById = new Map(($leaveTypes.data ?? []).map((t) => [t.id, t]));

	// Get badge color based on remaining days
	function getRemainingDaysColor(remaining: number, entitled: number): string {
		const percentage = entitled > 0 ? (remaining / entitled) * 100 : 0;
		if (percentage > 50) return 'border border-success/40 bg-success/15 text-foreground';
		if (percentage > 25) return 'border border-warning/50 bg-warning/15 text-foreground';
		return 'border border-destructive/40 bg-destructive/10 text-foreground';
	}

	// Used and pending as separate segments so the bar matches the legend (primary = used, warning = pending).
	function usageSegments(used: number, pending: number, entitled: number): [number, number] {
		if (entitled <= 0) return [0, 0];
		const usedPct = Math.min(100, (used / entitled) * 100);
		const pendingPct = Math.min(100 - usedPct, (pending / entitled) * 100);
		return [usedPct, pendingPct];
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
					{@const seg = usageSegments(typeBalance.used_days, typeBalance.pending_days, typeBalance.entitled_days)}
					<Card>
						<CardHeader class="pb-3">
							<CardTitle class="text-base">{typeById.get(typeBalance.leave_type_id)?.name ?? 'Leave'}</CardTitle>
							<CardDescription>{typeById.get(typeBalance.leave_type_id)?.code ?? ''}</CardDescription>
						</CardHeader>
						<CardContent class="space-y-4">
							<!-- Entitled Days -->
							<div>
								<p class="text-xs font-medium text-muted-foreground uppercase">Entitled</p>
								<p class="mt-1 text-2xl font-bold tabular-nums">{typeBalance.entitled_days}</p>
							</div>

							<!-- Usage Breakdown -->
							<div class="space-y-2 border-t pt-3">
								<div class="flex justify-between text-sm">
									<span class="text-muted-foreground">Used</span>
									<span class="font-medium tabular-nums">{typeBalance.used_days}</span>
								</div>
								<div class="flex justify-between text-sm">
									<span class="text-muted-foreground">Pending</span>
									<span class="font-medium tabular-nums">{typeBalance.pending_days}</span>
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
								<p class="text-lg tabular-nums">{typeBalance.remaining_days}</p>
							</div>

							<!-- Progress Bar -->
							<div class="space-y-1">
								<p class="text-xs text-muted-foreground">Usage</p>
								<div class="flex h-2 overflow-hidden rounded-full bg-muted" aria-hidden="true">
									<div class="h-full bg-primary transition-all" style="width: {seg[0]}%"></div>
									<div class="h-full bg-warning transition-all" style="width: {seg[1]}%"></div>
								</div>
								<p class="text-right text-xs tabular-nums text-muted-foreground">
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
					<span><strong>Pending:</strong> Requests awaiting approval</span>
				</div>
				<div class="flex items-center gap-2">
					<span class="inline-block h-3 w-3 rounded border border-border bg-surface"></span>
					<span><strong>Remaining:</strong> Available days to use</span>
				</div>
			</div>
		</div>
	{/if}
</div>

