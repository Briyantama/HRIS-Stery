<?php

namespace App\Providers;

use App\Services\LeaveServiceClient;
use Illuminate\Http\Request;
use Illuminate\Support\ServiceProvider;

class AppServiceProvider extends ServiceProvider
{
    public function register(): void
    {
        $this->app->bind(LeaveServiceClient::class, function ($app) {
            return new LeaveServiceClient($app->make(Request::class));
        });
    }

    public function boot(): void
    {
        //
    }
}
