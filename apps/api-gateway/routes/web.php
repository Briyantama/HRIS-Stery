<?php

use Illuminate\Support\Facades\Route;

Route::get('/', function () {
    return response()->json([
        'code' => 'OK',
        'message' => 'HRIS-Stery API Gateway',
    ]);
});
