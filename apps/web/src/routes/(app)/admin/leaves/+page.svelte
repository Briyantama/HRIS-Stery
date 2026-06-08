<script lang="ts">
	import { isManager, user } from '$lib/stores/auth';
	import { useLeaveRequests, useApproveLeave, useRejectLeave } from '$lib/queries/leave';
	import { RejectLeaveFormSchema, type RejectLeaveForm } from '$lib/schemas/leave';
	import { getErrorMessage, getValidationErrors, type ApiCallError } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import {
		Dialog,
		DialogContent,
		DialogDescription,
		DialogHeader,
		DialogTitle,
		DialogFooter
	} from '$lib/components/ui/dialog';
	import { AlertCircle, CheckCircle2, Loader2, X } from 'lucide-svelte';

	// Get current user
	let userId: string | null = null;
	user.subscribe((u) => {
		userId = u?.user_id ?? null;
	});

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
	 * Get badge color for status
	 */
	function getStatusBadgeColor(status: string): string {
		switch (status) {
			case 'PENDING':
				return 'bg-yellow-100 text-yellow-800';
			case 'APPROVED':
				return 'bg-green-100 text-green-800';
			case 'REJECTED':
				return 'bg-red-100 text-red-800';
			case 'CANCELLED':
				return 'bg-gray-100 text-gray-800';
			default:
				return 'bg-gray-100 text-gray-800';
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
		<Card class="border-red-200 bg-red-50">
			<CardHeader>
				<CardTitle class="flex items-center gap-2 text-red-800">
					<AlertCircle class="h-5 w-5" />
					Access Denied
				</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="text-sm text-red-700">
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
			<p class="mt-2 text-gray-600">Review and manage pending leave requests for your team</p>
		</div>

		<!-- Success Toast -->
		{#if showSuccessToast}
			<div class="fixed right-4 top-4 flex items-center gap-3 rounded-lg border border-green-200 bg-green-50 p-4 shadow-lg">
				<CheckCircle2 class="h-5 w-5 text-green-600" />
				<p class="text-sm font-medium text-green-800">{successMessage}</p>
				<button on:click={() => (showSuccessToast = false)} class="text-green-600 hover:text-green-700">
					<X class="h-4 w-4" />
				</button>
			</div>
		{/if}

		<!-- Status Filter Tabs -->
		<div class="flex gap-2 border-b border-gray-200">
			{#each [
				{ label: 'Pending', value: 'PENDING' as const },
				{ label: 'Approved', value: 'APPROVED' as const },
				{ label: 'Rejected', value: 'REJECTED' as const }
			] as tab}
				<button
					on:click={() => (statusFilter = tab.value)}
					class={`px-4 py-2 font-medium transition-colors ${
						statusFilter === tab.value
							? 'border-b-2 border-blue-500 text-blue-600'
							: 'text-gray-600 hover:text-gray-900'
					}`}
				>
					{tab.label}
				</button>
			{/each}
		</div>

		<!-- Loading State -->
		{#if $leaveRequests.isLoading}
			<div class="space-y-4">
				{#each Array(3) as _}
					<Card>
						<CardContent class="py-6">
							<div class="space-y-3">
								<div class="h-4 w-32 rounded bg-gray-200" />
								<div class="h-3 w-48 rounded bg-gray-100" />
								<div class="h-3 w-40 rounded bg-gray-100" />
							</div>
						</CardContent>
					</Card>
				{/each}
			</div>
		{:else if $leaveRequests.isError}
			<Card class="border-red-200 bg-red-50">
				<CardHeader>
					<CardTitle class="flex items-center gap-2 text-red-800">
						<AlertCircle class="h-5 w-5" />
						Error Loading Requests
					</CardTitle>
				</CardHeader>
				<CardContent>
					<p class="text-sm text-red-700">{getErrorMessage($leaveRequests.error)}</p>
				</CardContent>
			</Card>
		{:else if !$leaveRequests.data?.leave_requests || $leaveRequests.data.leave_requests.length === 0}
			<Card>
				<CardContent class="py-8 text-center">
					<p class="text-gray-600">No {statusFilter.toLowerCase()} leave requests to display</p>
				</CardContent>
			</Card>
		{:else}
			<!-- Leave Requests List -->
			<div class="space-y-4">
				{#each $leaveRequests.data.leave_requests as leave (leave.id)}
					<Card>
						<CardContent class="py-6">
							<!-- Request Header -->
							<div class="mb-4 flex flex-col items-start justify-between gap-3 md:flex-row md:items-center">
								<div>
									<div class="flex items-center gap-3">
										<h3 class="font-semibold text-gray-900">{leave.reason || 'Leave Request'}</h3>
										<span class={`rounded-full px-3 py-1 text-xs font-medium ${getStatusBadgeColor(leave.status)}`}>
											{leave.status}
										</span>
									</div>
									<p class="mt-1 text-sm text-gray-600">
										Requested by: <span class="font-medium">{leave.employee_id}</span>
									</p>
								</div>
							</div>

							<!-- Request Details Grid -->
							<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
								<!-- Dates -->
								<div>
									<p class="text-xs font-medium text-gray-500 uppercase">Dates</p>
									<p class="mt-1 text-sm font-medium text-gray-900">
										{formatDate(leave.start_date)} - {formatDate(leave.end_date)}
									</p>
									<p class="text-xs text-gray-600">
										({calculateDays(leave.start_date, leave.end_date)} days)
									</p>
								</div>

								<!-- Leave Type -->
								<div>
									<p class="text-xs font-medium text-gray-500 uppercase">Leave Type</p>
									<p class="mt-1 text-sm font-medium text-gray-900">{leave.leave_type_id}</p>
								</div>

								<!-- Status Info -->
								<div>
									<p class="text-xs font-medium text-gray-500 uppercase">
										{leave.status === 'APPROVED' ? 'Approved' : leave.status === 'REJECTED' ? 'Rejected' : 'Status'}
									</p>
									{#if leave.status === 'APPROVED' && leave.approved_at}
										<p class="mt-1 text-xs text-gray-600">on {formatDate(leave.approved_at)}</p>
										{#if leave.approved_by_id}
											<p class="text-xs text-gray-600">by {leave.approved_by_id}</p>
										{/if}
									{:else if leave.status === 'REJECTED' && leave.rejection_reason}
										<p class="mt-1 text-xs text-gray-600">{leave.rejection_reason}</p>
									{:else if leave.status === 'PENDING'}
										<p class="mt-1 text-xs text-yellow-600">Awaiting approval</p>
									{:else}
										<p class="mt-1 text-xs text-gray-600">—</p>
									{/if}
								</div>

								<!-- Actions (Only for Pending) -->
								{#if leave.status === 'PENDING'}
									<div class="flex gap-2 pt-4 md:col-span-2 lg:col-span-1">
										<Button
											size="sm"
											variant="default"
											on:click={() => handleApprove(leave.id)}
											disabled={$approveLeave.isPending}
											class="flex-1"
										>
											{#if $approveLeave.isPending}
												<Loader2 class="mr-1 h-3 w-3 animate-spin" />
											{/if}
											Approve
										</Button>
										<Button
											size="sm"
											variant="destructive"
											on:click={() => openRejectModal(leave.id)}
											disabled={$rejectLeave.isPending}
										>
											{#if $rejectLeave.isPending}
												<Loader2 class="mr-1 h-3 w-3 animate-spin" />
											{/if}
											Reject
										</Button>
									</div>
								{/if}
							</div>

							<!-- Reason (if provided) -->
							{#if leave.reason && leave.status === 'PENDING'}
								<div class="mt-4 border-t border-gray-200 pt-4">
									<p class="text-xs font-medium text-gray-500 uppercase">Employee Reason</p>
									<p class="mt-1 text-sm text-gray-700">{leave.reason}</p>
								</div>
							{/if}
						</CardContent>
					</Card>
				{/each}
			</div>

			<!-- Pagination Info -->
			{#if $leaveRequests.data.total_count > 0}
				<div class="text-center text-sm text-gray-600">
					Showing {$leaveRequests.data.leave_requests.length} of {$leaveRequests.data.total_count} requests
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
					<div class="flex items-center gap-3 rounded-lg border border-red-200 bg-red-50 p-3">
						<AlertCircle class="h-4 w-4 text-red-600" />
						<p class="text-sm text-red-700">{rejectGeneralError}</p>
					</div>
				{/if}

				<!-- Rejection Reason Textarea -->
				<div class="space-y-2">
					<label class="text-sm font-medium text-gray-900">Rejection Reason *</label>
					<Textarea
						placeholder="Explain why you are rejecting this leave request..."
						bind:value={rejectReason}
						disabled={$rejectLeave.isPending}
						class={`min-h-24 resize-none ${rejectErrors.reason ? 'border-red-500' : ''}`}
					/>
					{#if rejectErrors.reason}
						<p class="text-sm text-red-600">{rejectErrors.reason[0]}</p>
					{/if}
				</div>
			</div>

			<DialogFooter>
				<Button
					variant="outline"
					on:click={() => (showRejectModal = false)}
					disabled={$rejectLeave.isPending}
				>
					Cancel
				</Button>
				<Button
					variant="destructive"
					on:click={handleRejectSubmit}
					disabled={!rejectReason.trim() || $rejectLeave.isPending}
				>
					{#if $rejectLeave.isPending}
						<Loader2 class="mr-2 h-4 w-4 animate-spin" />
						Rejecting...
					{:else}
						Reject
					{/if}
				</Button>
			</DialogFooter>
		</DialogContent>
	</Dialog>
{/if}

<style>
	:global(body) {
		/* Ensure consistent spacing */
	}
</style>
