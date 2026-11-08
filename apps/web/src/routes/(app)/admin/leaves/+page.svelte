<script lang="ts">
	import { isManager } from '$lib/stores/auth';
	import { useLeaveRequests, useApproveLeave, useRejectLeave } from '$lib/queries/leave';
	import { RejectLeaveFormSchema } from '$lib/schemas/leave';
	import { getErrorMessage } from '$lib/api/client';
	import {
		Alert,
		Badge,
		Button,
		Card,
		CardContent,
		CardHeader,
		CardTitle,
		Dialog,
		DialogContent,
		DialogDescription,
		DialogFooter,
		DialogHeader,
		DialogTitle,
		Label,
		Skeleton,
		Textarea
	} from '@stery/kit/ui';
	import { CircleAlert, CircleCheck, X } from '@lucide/svelte';

	// Mutations
	const approveLeave = useApproveLeave();
	const rejectLeave = useRejectLeave();

	// Rejection modal state
	let showRejectModal = false;
	let selectedLeaveId: string | null = null;
	let rejectReason = '';
	let rejectErrors: Record<string, string[]> = {};
	let rejectGeneralError = '';

	// Status filter
	let statusFilter: 'PENDING' | 'APPROVED' | 'REJECTED' = 'PENDING';

	// Fetch leave requests based on status filter
	$: leaveRequests = useLeaveRequests({ status: statusFilter });

	// Toast state
	let showSuccessToast = false;
	let successMessage = '';

	/**
	 * Handle approve action with confirmation
	 */
	async function handleApprove(leaveId: string) {
		if (!confirm('Are you sure you want to approve this leave request?')) {
			return;
		}

		try {
			await $approveLeave.mutateAsync(leaveId);
			showSuccessToast = true;
			successMessage = 'Leave request approved successfully!';
			setTimeout(() => {
				showSuccessToast = false;
			}, 3000);

			// Refetch data
			$leaveRequests.refetch();
		} catch (error) {
			console.error('Approval failed:', error);
			alert(getErrorMessage(error));
		}
	}

	/**
	 * Open rejection modal for a specific leave request
	 */
	function openRejectModal(leaveId: string) {
		selectedLeaveId = leaveId;
		rejectReason = '';
		rejectErrors = {};
		rejectGeneralError = '';
		showRejectModal = true;
	}

	/**
	 * Handle rejection with reason validation
	 */
	async function handleRejectSubmit() {
		if (!selectedLeaveId) return;

		// Validate reason using Zod
		const result = RejectLeaveFormSchema.safeParse({ reason: rejectReason });

		if (!result.success) {
			rejectErrors = {};
			result.error.errors.forEach((error) => {
				const field = error.path[0] as string;
				if (field) {
					if (!rejectErrors[field]) {
						rejectErrors[field] = [];
					}
					rejectErrors[field].push(error.message);
				}
			});
			return;
		}

		try {
			await $rejectLeave.mutateAsync({
				leaveId: selectedLeaveId,
				reason: rejectReason
			});

			showSuccessToast = true;
			successMessage = 'Leave request rejected successfully!';
			setTimeout(() => {
				showSuccessToast = false;
			}, 3000);

			showRejectModal = false;
			selectedLeaveId = null;
			rejectReason = '';

			// Refetch data
			$leaveRequests.refetch();
		} catch (error) {
			rejectGeneralError = getErrorMessage(error);
		}
	}

	/**
	 * Format date for display
	 */
	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	/**
	 * Get badge tone for status
	 */
	function getStatusTone(status: string): 'warning' | 'success' | 'destructive' | 'neutral' {
		switch (status) {
			case 'PENDING':
				return 'warning';
			case 'APPROVED':
				return 'success';
			case 'REJECTED':
				return 'destructive';
			default:
				return 'neutral';
		}
	}

	/**
	 * Calculate days count
	 */
	function calculateDays(startDate: string, endDate: string): number {
		return (
			Math.ceil(
				(new Date(endDate).getTime() - new Date(startDate).getTime()) / (1000 * 60 * 60 * 24)
			) + 1
		);
	}
</script>

<!-- RBAC Guard: Only show to managers and HR admins -->
{#if !$isManager}
	<div class="mx-auto max-w-4xl py-6">
		<Card class="border-destructive/40 bg-destructive/10">
			<CardHeader>
				<CardTitle class="flex items-center gap-2 text-destructive">
					<CircleAlert class="h-5 w-5" />
					Access Denied
				</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="text-sm text-destructive">
					You do not have permission to access the leave approval portal. Only managers and HR administrators
					can view and manage leave requests.
				</p>
				<Button variant="outline" class="mt-4" href="/dashboard">Back to Dashboard</Button>
			</CardContent>
		</Card>
	</div>
{:else}
	<div class="space-y-6 py-6">
		<!-- Page Header -->
		<div>
			<h1 class="text-3xl font-bold tracking-tight">Leave Approval Portal</h1>
			<p class="mt-2 text-muted-foreground">Review and manage pending leave requests for your team</p>
		</div>

		<!-- Success Toast -->
		{#if showSuccessToast}
			<div class="fixed right-4 top-4 z-50 flex items-center gap-3 rounded-lg border border-success/40 bg-surface p-4 shadow-lg">
				<CircleCheck class="h-5 w-5 text-success" />
				<p class="text-sm font-medium text-foreground">{successMessage}</p>
				<button on:click={() => (showSuccessToast = false)} class="text-muted-foreground hover:text-foreground" aria-label="Dismiss">
					<X class="h-4 w-4" />
				</button>
			</div>
		{/if}

		<!-- Status Filter Tabs -->
		<div class="flex gap-2 border-b border-border">
			{#each [
				{ label: 'Pending', value: 'PENDING' as const },
				{ label: 'Approved', value: 'APPROVED' as const },
				{ label: 'Rejected', value: 'REJECTED' as const }
			] as tab}
				<button
					on:click={() => (statusFilter = tab.value)}
					class={`px-4 py-2 font-medium transition-colors ${
						statusFilter === tab.value
							? 'border-b-2 border-primary text-primary'
							: 'text-muted-foreground hover:text-foreground'
					}`}
				>
					{tab.label}
				</button>
			{/each}
		</div>

		<!-- Loading State -->
		{#if $leaveRequests.isLoading}
			<div class="space-y-4">
				{#each [0, 1, 2] as i (i)}
					<Card>
						<CardContent class="py-6">
							<div class="space-y-3">
								<Skeleton class="h-4 w-32" />
								<Skeleton class="h-3 w-48" />
								<Skeleton class="h-3 w-40" />
							</div>
						</CardContent>
					</Card>
				{/each}
			</div>
		{:else if $leaveRequests.isError}
			<Alert tone="destructive" title="Error Loading Requests" description={getErrorMessage($leaveRequests.error)}>
				{#snippet icon()}<CircleAlert class="h-5 w-5 text-destructive" />{/snippet}
			</Alert>
		{:else if !$leaveRequests.data?.items || $leaveRequests.data.items.length === 0}
			<Card>
				<CardContent class="py-8 text-center">
					<p class="text-muted-foreground">No {statusFilter.toLowerCase()} leave requests to display</p>
				</CardContent>
			</Card>
		{:else}
			<!-- Leave Requests List -->
			<div class="space-y-4">
				{#each $leaveRequests.data.items as leave (leave.id)}
					<Card>
						<CardContent class="py-6">
							<!-- Request Header -->
							<div class="mb-4 flex flex-col items-start justify-between gap-3 md:flex-row md:items-center">
								<div>
									<div class="flex items-center gap-3">
										<h3 class="font-semibold text-foreground">{leave.reason || 'Leave Request'}</h3>
										<Badge tone={getStatusTone(leave.status)} class="rounded-full px-3 py-1">
											{leave.status}
										</Badge>
									</div>
									<p class="mt-1 text-sm text-muted-foreground">
										Requested by: <span class="font-medium">{leave.employee_id}</span>
									</p>
								</div>
							</div>

							<!-- Request Details Grid -->
							<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
								<!-- Dates -->
								<div>
									<p class="text-xs font-medium text-muted-foreground uppercase">Dates</p>
									<p class="mt-1 text-sm font-medium text-foreground">
										{formatDate(leave.start_date)} - {formatDate(leave.end_date)}
									</p>
									<p class="text-xs text-muted-foreground">
										({calculateDays(leave.start_date, leave.end_date)} days)
									</p>
								</div>

								<!-- Leave Type -->
								<div>
									<p class="text-xs font-medium text-muted-foreground uppercase">Leave Type</p>
									<p class="mt-1 text-sm font-medium text-foreground">{leave.leave_type_id}</p>
								</div>

								<!-- Status Info -->
								<div>
									<p class="text-xs font-medium text-muted-foreground uppercase">
										{leave.status === 'APPROVED' ? 'Approved' : leave.status === 'REJECTED' ? 'Rejected' : 'Status'}
									</p>
									{#if leave.status === 'APPROVED' && leave.approved_at}
										<p class="mt-1 text-xs text-muted-foreground">on {formatDate(leave.approved_at)}</p>
										{#if leave.approved_by_id}
											<p class="text-xs text-muted-foreground">by {leave.approved_by_id}</p>
										{/if}
									{:else if leave.status === 'REJECTED' && leave.rejection_reason}
										<p class="mt-1 text-xs text-muted-foreground">{leave.rejection_reason}</p>
									{:else if leave.status === 'PENDING'}
										<p class="mt-1 text-xs text-warning">Awaiting approval</p>
									{:else}
										<p class="mt-1 text-xs text-muted-foreground">—</p>
									{/if}
								</div>

								<!-- Actions (Only for Pending) -->
								{#if leave.status === 'PENDING'}
									<div class="flex gap-2 pt-4 md:col-span-2 lg:col-span-1">
										<Button
											size="sm"
											variant="primary"
											onclick={() => handleApprove(leave.id)}
											loading={$approveLeave.isPending}
											class="flex-1"
										>
											Approve
										</Button>
										<Button
											size="sm"
											variant="destructive"
											onclick={() => openRejectModal(leave.id)}
											loading={$rejectLeave.isPending}
										>
											Reject
										</Button>
									</div>
								{/if}
							</div>

							<!-- Reason (if provided) -->
							{#if leave.reason && leave.status === 'PENDING'}
								<div class="mt-4 border-t border-border pt-4">
									<p class="text-xs font-medium text-muted-foreground uppercase">Employee Reason</p>
									<p class="mt-1 text-sm text-foreground">{leave.reason}</p>
								</div>
							{/if}
						</CardContent>
					</Card>
				{/each}
			</div>

			<!-- Pagination Info -->
			{#if $leaveRequests.data.total > 0}
				<div class="text-center text-sm text-muted-foreground">
					Showing {$leaveRequests.data.items.length} of {$leaveRequests.data.total} requests
				</div>
			{/if}
		{/if}
	</div>

	<!-- Rejection Modal -->
	<Dialog bind:open={showRejectModal}>
		<DialogContent class="max-w-md">
			<DialogHeader>
				<DialogTitle>Reject Leave Request</DialogTitle>
				<DialogDescription>Provide a reason for rejecting this leave request</DialogDescription>
			</DialogHeader>

			<div class="space-y-4 py-4">
				<!-- General Error -->
				{#if rejectGeneralError}
					<Alert tone="destructive">{rejectGeneralError}</Alert>
				{/if}

				<!-- Rejection Reason Textarea -->
				<div class="space-y-2">
					<Label for="reject-reason">Rejection Reason *</Label>
					<Textarea
						id="reject-reason"
						placeholder="Explain why you are rejecting this leave request..."
						bind:value={rejectReason}
						disabled={$rejectLeave.isPending}
						invalid={!!rejectErrors.reason}
						class="min-h-24 resize-none"
					/>
					{#if rejectErrors.reason}
						<p class="text-sm text-destructive">{rejectErrors.reason[0]}</p>
					{/if}
				</div>
			</div>

			<DialogFooter>
				<Button
					variant="outline"
					onclick={() => (showRejectModal = false)}
					disabled={$rejectLeave.isPending}
				>
					Cancel
				</Button>
				<Button
					variant="destructive"
					onclick={handleRejectSubmit}
					disabled={!rejectReason.trim()}
					loading={$rejectLeave.isPending}
				>
					{$rejectLeave.isPending ? 'Rejecting...' : 'Reject'}
				</Button>
			</DialogFooter>
		</DialogContent>
	</Dialog>
{/if}
