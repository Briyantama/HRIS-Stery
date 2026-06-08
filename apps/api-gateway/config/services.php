<?php

return [
    'leave' => [
        'url' => env('LEAVE_SERVICE_URL', 'http://localhost:8084'),
    ],
    'employee' => [
        'url' => env('EMPLOYEE_SERVICE_URL', 'http://localhost:8082'),
    ],
    'attendance' => [
        'url' => env('ATTENDANCE_SERVICE_URL', 'http://localhost:8083'),
    ],
    'auth' => [
        'url' => env('AUTH_SERVICE_URL', 'http://localhost:8081'),
    ],
];
