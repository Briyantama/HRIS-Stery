-- Create notification schema
CREATE SCHEMA IF NOT EXISTS notification;

-- Enable RLS on the schema
ALTER DEFAULT PRIVILEGES IN SCHEMA notification GRANT ALL ON TABLES TO app;
ALTER DEFAULT PRIVILEGES IN SCHEMA notification GRANT ALL ON SEQUENCES TO app;

-- Create notifications table with RLS
CREATE TABLE notification.notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    recipient_id UUID NOT NULL,
    channel VARCHAR(20) NOT NULL CHECK (channel IN ('EMAIL', 'IN_APP')),
    template_key VARCHAR(255) NOT NULL,
    variables JSONB,
    subject VARCHAR(500),
    body TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SENT', 'FAILED', 'READ')),
    delivery_id VARCHAR(255),
    delivery_error TEXT,
    scheduled_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Enable RLS on notifications
ALTER TABLE notification.notifications ENABLE ROW LEVEL SECURITY;
CREATE POLICY notifications_tenant_isolation ON notification.notifications
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
CREATE POLICY notifications_tenant_isolation_insert ON notification.notifications
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

-- Create notification_channel_configs table with RLS
CREATE TABLE notification.notification_channel_configs (
    tenant_id UUID PRIMARY KEY,
    channels TEXT[] NOT NULL,
    email_config JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Enable RLS on notification_channel_configs
ALTER TABLE notification.notification_channel_configs ENABLE ROW LEVEL SECURITY;
CREATE POLICY channel_config_tenant_isolation ON notification.notification_channel_configs
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
CREATE POLICY channel_config_tenant_isolation_insert ON notification.notification_channel_configs
    WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

-- Create processed_events table (NO RLS - internal idempotency)
CREATE TABLE notification.processed_events (
    event_id VARCHAR(255) PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create indexes
CREATE INDEX idx_notifications_tenant_id ON notification.notifications(tenant_id);
CREATE INDEX idx_notifications_tenant_recipient ON notification.notifications(tenant_id, recipient_id);
CREATE INDEX idx_notifications_tenant_status ON notification.notifications(tenant_id, status);
CREATE INDEX idx_notifications_tenant_created ON notification.notifications(tenant_id, created_at DESC);
CREATE INDEX idx_notifications_recipient_read ON notification.notifications(recipient_id, read_at);
CREATE INDEX idx_processed_events_event_id ON notification.processed_events(event_id);

-- Grant permissions to app role
GRANT USAGE ON SCHEMA notification TO app;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA notification TO app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA notification TO app;
