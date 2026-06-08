import { auth, isAuthenticated } from '$lib/stores/auth';
import { goto } from '$app/navigation';
import type { ApiError } from '$lib/schemas/leave';

/**
 * API Client wrapper around fetch API.
 *
 * Features:
 * - Automatically injects Authorization: Bearer <token> header
 * - Handles 401 Unauthorized by clearing session and redirecting to login
 * - Standardized error handling with typed responses
 * - Request ID generation for distributed tracing
 */

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8000/api/v1';

interface FetchOptions extends RequestInit {
	query?: Record<string, string | number | boolean | undefined>;
}

interface ApiResponse<T> {
	code: string;
	message?: string;
	data?: T;
	details?: string[];
}

/**
 * Internal function to extract auth token from store.
 * Used to get current token at request time (reactive).
 */
let currentToken: string | null = null;
auth.subscribe((state) => {
	currentToken = state.token;
});

/**
 * Make authenticated API request with automatic token injection.
 * Handles errors and 401 redirects.
 *
 * @param method HTTP method (GET, POST, PUT, DELETE, etc.)
 * @param path API endpoint path (e.g., '/leaves')
 * @param options Fetch options (body, query params, headers, etc.)
 * @returns Parsed JSON response
 * @throws ApiError with code and message
 */
export async function apiCall<T = unknown>(
	method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH',
	path: string,
	options: FetchOptions = {}
): Promise<ApiResponse<T>> {
	// Build URL with query parameters
	let url = `${API_BASE_URL}${path}`;
	if (options.query) {
		const params = new URLSearchParams();
		Object.entries(options.query).forEach(([key, value]) => {
			if (value !== undefined && value !== null) {
				params.append(key, String(value));
			}
		});
		const queryString = params.toString();
		if (queryString) {
			url += `?${queryString}`;
		}
	}

	// Prepare headers
	const headers: HeadersInit = {
		'Content-Type': 'application/json',
		...options.headers
	};

	// Inject authorization token
	if (currentToken) {
		headers['Authorization'] = `Bearer ${currentToken}`;
	}

	// Generate request ID for tracing
	headers['X-Request-ID'] = generateRequestId();

	// Make request
	const response = await fetch(url, {
		method,
		...options,
		headers
	});

	// Handle 401 Unauthorized - clear session and redirect
	if (response.status === 401) {
		auth.clearSession();
		if (typeof window !== 'undefined') {
			await goto('/login');
		}
		throw new Error('Unauthorized');
	}

	// Parse response
	let data: ApiResponse<T>;
	try {
		data = await response.json();
	} catch (e) {
		// Response is not JSON
		throw new ApiCallError(
			response.status === 500 ? 'INTERNAL' : 'UNKNOWN',
			`HTTP ${response.status}: ${response.statusText}`,
			response.status
		);
	}

	// Handle non-2xx responses
	if (!response.ok) {
		const error = data as ApiError;
		throw new ApiCallError(error.code || 'UNKNOWN', error.message || 'Unknown error', response.status, error.details);
	}

	return data;
}

/**
 * GET request
 */
export function apiGet<T = unknown>(
	path: string,
	options: FetchOptions = {}
): Promise<ApiResponse<T>> {
	return apiCall<T>('GET', path, { ...options, method: 'GET' });
}

/**
 * POST request
 */
export function apiPost<T = unknown>(
	path: string,
	body?: unknown,
	options: FetchOptions = {}
): Promise<ApiResponse<T>> {
	return apiCall<T>('POST', path, {
		...options,
		method: 'POST',
		body: body ? JSON.stringify(body) : undefined
	});
}

/**
 * PUT request
 */
export function apiPut<T = unknown>(
	path: string,
	body?: unknown,
	options: FetchOptions = {}
): Promise<ApiResponse<T>> {
	return apiCall<T>('PUT', path, {
		...options,
		method: 'PUT',
		body: body ? JSON.stringify(body) : undefined
	});
}

/**
 * DELETE request
 */
export function apiDelete<T = unknown>(
	path: string,
	options: FetchOptions = {}
): Promise<ApiResponse<T>> {
	return apiCall<T>('DELETE', path, { ...options, method: 'DELETE' });
}

/**
 * Custom error class for API errors
 */
export class ApiCallError extends Error {
	constructor(
		public code: string,
		message: string,
		public status: number = 500,
		public details?: string[]
	) {
		super(message);
		this.name = 'ApiCallError';
	}
}

/**
 * Generate unique request ID for distributed tracing
 */
function generateRequestId(): string {
	return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

/**
 * Parse error response and return user-friendly message
 */
export function getErrorMessage(error: unknown): string {
	if (error instanceof ApiCallError) {
		return error.message || `Error (${error.code})`;
	}
	if (error instanceof Error) {
		return error.message;
	}
	return 'An unexpected error occurred';
}

/**
 * Check if error is 401 Unauthorized
 */
export function isUnauthorized(error: unknown): boolean {
	return error instanceof ApiCallError && error.status === 401;
}

/**
 * Check if error is 403 Forbidden
 */
export function isForbidden(error: unknown): boolean {
	return error instanceof ApiCallError && error.status === 403;
}

/**
 * Check if error is 404 Not Found
 */
export function isNotFound(error: unknown): boolean {
	return error instanceof ApiCallError && error.status === 404;
}

/**
 * Extract validation errors from API response
 */
export function getValidationErrors(error: unknown): Record<string, string[]> {
	if (error instanceof ApiCallError && error.details) {
		// Parse details array into field-level errors
		// Assumes details are in format: "field: error message"
		const errors: Record<string, string[]> = {};
		error.details.forEach((detail) => {
			const parts = detail.split(':');
			if (parts.length > 1) {
				const field = parts[0].trim();
				const message = parts.slice(1).join(':').trim();
				if (!errors[field]) {
					errors[field] = [];
				}
				errors[field].push(message);
			} else {
				// If no field specified, add to generic errors
				if (!errors['_general']) {
					errors['_general'] = [];
				}
				errors['_general'].push(detail);
			}
		});
		return errors;
	}
	return {};
}
