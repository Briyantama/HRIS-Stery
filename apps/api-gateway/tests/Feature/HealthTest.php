<?php

it('returns gateway health status', function () {
    $response = $this->getJson('/api/health');

    $response->assertOk()
        ->assertJson([
            'code' => 'OK',
            'message' => 'Gateway is healthy',
        ]);
});
