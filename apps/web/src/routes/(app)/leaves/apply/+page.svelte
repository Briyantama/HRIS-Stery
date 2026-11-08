<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { useLeaveTypes, useApplyLeave } from '$lib/queries/leave';
	import { ApplyLeaveRequestSchema, type ApplyLeaveForm } from '$lib/schemas/leave';
	import { getErrorMessage } from '$lib/api/client';
	import {
		Alert,
		Button,
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle,
		Input,
		Label,
		Select,
		Textarea
	} from '@stery/kit/ui';
	import { CircleAlert, CircleCheck } from '@lucide/svelte';

	// Fetch leave types for dropdown
	const leaveTypes = useLeaveTypes();

	// Mutation hook for form submission
	const applyLeave = useApplyLeave();

	// Form state
	let formData: ApplyLeaveForm = {
		leave_type_id: '',
		start_date: '',
		end_date: '',
		reason: ''
	};

	// Validation state
	let fieldErrors: Record<string, string[]> = {};
	let generalError = '';
	let showSuccess = false;

	/**
	 * Validate form data using Zod schema
	 * Returns true if valid, false if validation errors exist
	 */
	function validateForm(): boolean {
		fieldErrors = {};
		generalError = '';

		const result = ApplyLeaveRequestSchema.safeParse(formData);

		if (!result.success) {
			// Map Zod errors to field-level errors
			result.error.errors.forEach((error) => {
				const field = error.path[0] as string;
				if (field) {
					if (!fieldErrors[field]) {
						fieldErrors[field] = [];
					}
					fieldErrors[field].push(error.message);
				}
			});
			return false;
		}

		return true;
	}

	/**
	 * Handle real-time validation on field change
	 */
	function handleFieldChange(field: keyof ApplyLeaveForm) {
		// Validate just this field
		if (field === 'end_date' && formData.start_date) {
			// Validate end_date >= start_date
			if (formData.end_date && formData.end_date < formData.start_date) {
				fieldErrors.end_date = ['End date must be on or after start date'];
			} else {
				delete fieldErrors.end_date;
			}
		}

		fieldErrors = fieldErrors; // Trigger reactivity
	}

	/**
	 * Handle form submission
	 */
	async function handleSubmit(e: Event) {
		e.preventDefault();

		// Validate entire form
		if (!validateForm()) {
			return;
		}

		generalError = '';
		showSuccess = false;

		try {
			// Submit via mutation
			await $applyLeave.mutateAsync(formData);

			// Success
			showSuccess = true;

			// Reset form
			formData = {
				leave_type_id: '',
				start_date: '',
				end_date: '',
				reason: ''
			};
			fieldErrors = {};

			// Redirect to dashboard after a brief moment
			setTimeout(() => {
				goto(resolve('/dashboard'));
			}, 1500);
		} catch (error) {
			generalError = getErrorMessage(error);
		}
	}

	/**
	 * Helper to check if a field has errors
	 */
	function hasError(field: keyof ApplyLeaveForm): boolean {
		return !!fieldErrors[field];
	}

	/**
	 * Get first error message for a field
	 */
	function getFieldError(field: keyof ApplyLeaveForm): string {
		return fieldErrors[field]?.[0] ?? '';
	}

	/**
	 * Check if form is valid for submission
	 */
	function isFormValid(): boolean {
		return !!(
			formData.leave_type_id &&
			formData.start_date &&
			formData.end_date &&
			Object.keys(fieldErrors).length === 0
		);
	}
</script>

<div class="mx-auto max-w-2xl space-y-6 py-6">
	<!-- Page Header -->
	<div>
		<h1 class="text-3xl font-bold tracking-tight">Request Leave</h1>
		<p class="mt-2 text-muted-foreground">Submit a new leave request to your manager</p>
	</div>

	<Card>
		<CardHeader>
			<CardTitle>Leave Request Details</CardTitle>
			<CardDescription>Fill in your leave request information below</CardDescription>
		</CardHeader>

		<CardContent>
			<form on:submit={handleSubmit} class="space-y-6">
				<!-- Success Message -->
				{#if showSuccess}
					<Alert tone="success" title="Leave request submitted successfully!" description="Redirecting to dashboard...">
						{#snippet icon()}<CircleCheck class="h-5 w-5 text-success" />{/snippet}
					</Alert>
				{/if}

				<!-- General Error Message -->
				{#if generalError}
					<Alert tone="destructive" title="Error submitting leave request" description={generalError}>
						{#snippet icon()}<CircleAlert class="h-5 w-5 text-destructive" />{/snippet}
					</Alert>
				{/if}

				<!-- Leave Type Selection -->
				<div class="space-y-2">
					<Label for="leave-type">Leave Type *</Label>

					{#if $leaveTypes.isLoading}
						<div class="h-10 rounded-lg border border-input bg-muted"></div>
					{:else if $leaveTypes.isError}
						<Alert tone="destructive">
							Failed to load leave types: {$leaveTypes.error?.message || 'Unknown error'}
						</Alert>
					{:else}
						<Select
							id="leave-type"
							items={($leaveTypes.data ?? []).map((t) => ({ value: t.id, label: t.name }))}
							value={formData.leave_type_id}
							onValueChange={(v: string) => (formData.leave_type_id = v)}
							placeholder="Select a leave type..."
							invalid={hasError('leave_type_id')}
						/>

						{#if hasError('leave_type_id')}
							<p class="text-sm text-destructive">{getFieldError('leave_type_id')}</p>
						{/if}
					{/if}
				</div>

				<!-- Date Range -->
				<div class="grid gap-4 md:grid-cols-2">
					<!-- Start Date -->
					<div class="space-y-2">
						<Label for="start-date">Start Date *</Label>
						<Input
							id="start-date"
							type="date"
							bind:value={formData.start_date}
							onchange={() => handleFieldChange('start_date')}
							disabled={$applyLeave.isPending}
							invalid={hasError('start_date')}
						/>
						{#if hasError('start_date')}
							<p class="text-sm text-destructive">{getFieldError('start_date')}</p>
						{/if}
					</div>

					<!-- End Date -->
					<div class="space-y-2">
						<Label for="end-date">End Date *</Label>
						<Input
							id="end-date"
							type="date"
							bind:value={formData.end_date}
							onchange={() => handleFieldChange('end_date')}
							disabled={$applyLeave.isPending}
							invalid={hasError('end_date')}
						/>
						{#if hasError('end_date')}
							<p class="text-sm text-destructive">{getFieldError('end_date')}</p>
						{/if}
					</div>
				</div>

				<!-- Days Count Display -->
				{#if formData.start_date && formData.end_date}
					{@const daysCount = Math.ceil(
						(new Date(formData.end_date).getTime() - new Date(formData.start_date).getTime()) /
							(1000 * 60 * 60 * 24)
					) + 1}
					<div class="rounded-lg border border-info/40 bg-info/10 p-3">
						<p class="text-sm font-medium text-foreground">
							Total Leave Days: <span class="text-lg font-bold">{daysCount}</span>
						</p>
					</div>
				{/if}

				<!-- Reason -->
				<div class="space-y-2">
					<Label for="reason">Reason (Optional)</Label>
					<Textarea
						id="reason"
						placeholder="Enter your reason for leave (e.g., vacation, personal, medical)..."
						bind:value={formData.reason}
						disabled={$applyLeave.isPending}
						class="min-h-24 resize-none"
					/>
					{#if hasError('reason')}
						<p class="text-sm text-destructive">{getFieldError('reason')}</p>
					{/if}
				</div>

				<!-- Form Actions -->
				<div class="flex gap-3 pt-4">
					<Button
						type="submit"
						disabled={!isFormValid() || $applyLeave.isPending}
						class="flex-1"
					>
						{$applyLeave.isPending ? 'Submitting...' : 'Submit Request'}
					</Button>

					<Button
						type="button"
						variant="outline"
						onclick={() => goto(resolve('/dashboard'))}
						disabled={$applyLeave.isPending}
					>
						Cancel
					</Button>
				</div>
			</form>
		</CardContent>
	</Card>

	<!-- Help Text -->
	<div class="rounded-lg border border-border bg-muted p-4">
		<p class="text-sm font-medium text-foreground">Need Help?</p>
		<ul class="mt-2 space-y-1 text-sm text-muted-foreground">
			<li>• Select your leave type from the dropdown (Annual, Sick, etc.)</li>
			<li>• Choose your start and end dates (end date must be on or after start date)</li>
			<li>• Optionally provide a reason for your leave</li>
			<li>• Click "Submit Request" to send your request to your manager for approval</li>
		</ul>
	</div>
</div>
