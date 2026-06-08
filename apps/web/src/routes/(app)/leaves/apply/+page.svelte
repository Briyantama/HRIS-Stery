<script lang="ts">
	import { goto } from '$app/navigation';
	import { user } from '$lib/stores/auth';
	import { useLeaveTypes, useApplyLeave } from '$lib/queries/leave';
	import { ApplyLeaveRequestSchema, type ApplyLeaveForm } from '$lib/schemas/leave';
	import { getErrorMessage, getValidationErrors, type ApiCallError } from '$lib/api/client';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import {
		Select,
		SelectContent,
		SelectItem,
		SelectTrigger,
		SelectValue
	} from '$lib/components/ui/select';
	import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { AlertCircle, CheckCircle2, Loader2 } from 'lucide-svelte';

	// Get current user
	let userId: string | null = null;
	user.subscribe((u) => {
		userId = u?.user_id ?? null;
	});

	// Fetch leave types for dropdown
	const leaveTypes = useLeaveTypes();

	// Mutation hook for form submission
	const applyLeave = useApplyLeave();

	// Form state
	let formData: ApplyLeaveForm = {
		employee_id: userId ?? '',
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
		// Update employee_id from store if not set
		if (field === 'employee_id' && !formData.employee_id && userId) {
			formData.employee_id = userId;
		}

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
			// Ensure employee_id is set
			const submitData: ApplyLeaveForm = {
				...formData,
				employee_id: formData.employee_id || userId || ''
			};

			// Submit via mutation
			await $applyLeave.mutateAsync(submitData);

			// Success
			showSuccess = true;

			// Reset form
			formData = {
				employee_id: userId ?? '',
				leave_type_id: '',
				start_date: '',
				end_date: '',
				reason: ''
			};
			fieldErrors = {};

			// Redirect to dashboard after a brief moment
			setTimeout(() => {
				goto('/dashboard');
			}, 1500);
		} catch (error) {
			generalError = getErrorMessage(error);

			// Parse field-level validation errors from backend
			const backendErrors = getValidationErrors(error);
			if (Object.keys(backendErrors).length > 0) {
				fieldErrors = backendErrors;
			}
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
		return (
			formData.employee_id &&
			formData.leave_type_id &&
			formData.start_date &&
			formData.end_date &&
			Object.keys(fieldErrors).length === 0
		);
	}

	// Update employee_id when userId changes
	$: if (userId && !formData.employee_id) {
		formData.employee_id = userId;
	}
</script>

<div class="mx-auto max-w-2xl space-y-6 py-6">
	<!-- Page Header -->
	<div>
		<h1 class="text-3xl font-bold tracking-tight">Request Leave</h1>
		<p class="mt-2 text-gray-600">Submit a new leave request to your manager</p>
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
					<div class="flex items-center gap-3 rounded-lg border border-green-200 bg-green-50 p-4">
						<CheckCircle2 class="h-5 w-5 text-green-600" />
						<div>
							<p class="font-medium text-green-800">Leave request submitted successfully!</p>
							<p class="text-sm text-green-700">Redirecting to dashboard...</p>
						</div>
					</div>
				{/if}

				<!-- General Error Message -->
				{#if generalError}
					<div class="flex items-center gap-3 rounded-lg border border-red-200 bg-red-50 p-4">
						<AlertCircle class="h-5 w-5 text-red-600" />
						<div>
							<p class="font-medium text-red-800">Error submitting leave request</p>
							<p class="text-sm text-red-700">{generalError}</p>
						</div>
					</div>
				{/if}

				<!-- Leave Type Selection -->
				<div class="space-y-2">
					<Label for="leave-type">Leave Type *</Label>

					{#if $leaveTypes.isLoading}
						<div class="h-10 rounded-lg border border-gray-300 bg-gray-50" />
					{:else if $leaveTypes.isError}
						<div class="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
							Failed to load leave types: {$leaveTypes.error?.message || 'Unknown error'}
						</div>
					{:else}
						<Select value={formData.leave_type_id} onValueChange={(v) => (formData.leave_type_id = v)}>
							<SelectTrigger id="leave-type" class={hasError('leave_type_id') ? 'border-red-500' : ''}>
								<SelectValue placeholder="Select a leave type..." />
							</SelectTrigger>
							<SelectContent>
								{#each $leaveTypes.data ?? [] as type (type.id)}
									<SelectItem value={type.id}>{type.name}</SelectItem>
								{/each}
							</SelectContent>
						</Select>

						{#if hasError('leave_type_id')}
							<p class="text-sm text-red-600">{getFieldError('leave_type_id')}</p>
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
							on:change={() => handleFieldChange('start_date')}
							disabled={$applyLeave.isPending}
							class={hasError('start_date') ? 'border-red-500' : ''}
						/>
						{#if hasError('start_date')}
							<p class="text-sm text-red-600">{getFieldError('start_date')}</p>
						{/if}
					</div>

					<!-- End Date -->
					<div class="space-y-2">
						<Label for="end-date">End Date *</Label>
						<Input
							id="end-date"
							type="date"
							bind:value={formData.end_date}
							on:change={() => handleFieldChange('end_date')}
							disabled={$applyLeave.isPending}
							class={hasError('end_date') ? 'border-red-500' : ''}
						/>
						{#if hasError('end_date')}
							<p class="text-sm text-red-600">{getFieldError('end_date')}</p>
						{/if}
					</div>
				</div>

				<!-- Days Count Display -->
				{#if formData.start_date && formData.end_date}
					{@const daysCount = Math.ceil(
						(new Date(formData.end_date).getTime() - new Date(formData.start_date).getTime()) /
							(1000 * 60 * 60 * 24)
					) + 1}
					<div class="rounded-lg border border-blue-200 bg-blue-50 p-3">
						<p class="text-sm font-medium text-blue-900">
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
						<p class="text-sm text-red-600">{getFieldError('reason')}</p>
					{/if}
				</div>

				<!-- Form Actions -->
				<div class="flex gap-3 pt-4">
					<Button
						type="submit"
						disabled={!isFormValid() || $applyLeave.isPending}
						class="flex-1"
					>
						{#if $applyLeave.isPending}
							<Loader2 class="mr-2 h-4 w-4 animate-spin" />
							Submitting...
						{:else}
							Submit Request
						{/if}
					</Button>

					<Button
						type="button"
						variant="outline"
						on:click={() => goto('/dashboard')}
						disabled={$applyLeave.isPending}
					>
						Cancel
					</Button>
				</div>
			</form>
		</CardContent>
	</Card>

	<!-- Help Text -->
	<div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
		<p class="text-sm font-medium text-gray-700">Need Help?</p>
		<ul class="mt-2 space-y-1 text-sm text-gray-600">
			<li>• Select your leave type from the dropdown (Annual, Sick, etc.)</li>
			<li>• Choose your start and end dates (end date must be on or after start date)</li>
			<li>• Optionally provide a reason for your leave</li>
			<li>• Click "Submit Request" to send your request to your manager for approval</li>
		</ul>
	</div>
</div>
