<?php

use App\Http\Controllers\Api\V1\LeaveController;
use Illuminate\Support\Facades\Route;

/*
|--------------------------------------------------------------------------
| API Routes
|--------------------------------------------------------------------------
|
| Here is where you can register API routes for your application. These
| routes are loaded by the RouteServiceProvider and all of them will
| be assigned to the "api" middleware group.
|
*/

// Public routes (no JWT validation required)
Route::prefix('auth')->group(function () {
    // Auth routes will be implemented in AuthController
    // Route::post('login', [AuthController::class, 'login']);
    // Route::post('register', [AuthController::class, 'register']);
    // Route::post('refresh', [AuthController::class, 'refresh']);
});

// Health check endpoint (no authentication required)
Route::get('health', function () {
    return response()->json([
        'code' => 'OK',
        'message' => 'Gateway is healthy',
        'timestamp' => now()->toIso8601String(),
    ]);
});

// Protected routes (JWT validation + Tenant enforcement required)
Route::middleware([
    'api',
    \App\Http\Middleware\JwtValidation::class,
    \App\Http\Middleware\TenantResolver::class,
])->group(function () {
    // API v1 routes
    Route::prefix('v1')->group(function () {
        /*
        |--------------------------------------------------------------------------
        | Leave Management Routes
        |--------------------------------------------------------------------------
        |
        | These routes handle leave request workflows:
        | - Creation, retrieval, listing
        | - Approval and rejection (manager-only)
        | - Cancellation
        | - Balance and leave type queries
        |
        */
        Route::prefix('leaves')->group(function () {
            // Apply for new leave
            Route::post('/', [LeaveController::class, 'applyLeave'])
                ->name('leaves.apply');

            // List leave requests with optional filters
            Route::get('/', [LeaveController::class, 'listLeaveRequests'])
                ->name('leaves.list');

            // Get single leave request by ID
            Route::get('{id}', [LeaveController::class, 'getLeaveRequest'])
                ->name('leaves.show');

            // Approve leave request (manager-only)
            Route::post('{id}/approve', [LeaveController::class, 'approveLeave'])
                ->name('leaves.approve');

            // Reject leave request (manager-only)
            Route::post('{id}/reject', [LeaveController::class, 'rejectLeave'])
                ->name('leaves.reject');

            // Cancel leave request
            Route::post('{id}/cancel', [LeaveController::class, 'cancelLeave'])
                ->name('leaves.cancel');
        });

        // Get leave balance for employee
        Route::get('leave-balance/{employeeId}', [LeaveController::class, 'getLeaveBalance'])
            ->name('leave.balance');

        // List all leave types
        Route::get('leave-types', [LeaveController::class, 'listLeaveTypes'])
            ->name('leave.types');

        /*
        |--------------------------------------------------------------------------
        | Future Endpoints (Placeholder Routes)
        |--------------------------------------------------------------------------
        |
        | Additional endpoints for other domains will be added here:
        | - /employees   (Employee management)
        | - /attendance  (Attendance tracking)
        | - /documents   (Document uploads)
        | - /audit       (Audit logs)
        |
        */
    });
});

/*
|--------------------------------------------------------------------------
| Catch-all for undefined API routes
|--------------------------------------------------------------------------
|
| Return 404 for any undefined API endpoint.
|
*/
Route::fallback(function () {
    return response()->json([
        'code' => 'NOT_FOUND',
        'message' => 'API endpoint not found',
    ], 404);
});
