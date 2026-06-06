-- Create document schema
CREATE SCHEMA IF NOT EXISTS document;

-- Document files table
CREATE TABLE document.files (
  file_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL,
  uploaded_by UUID NOT NULL,
  entity_type VARCHAR(50) NOT NULL,
  entity_id UUID NOT NULL,

  file_name VARCHAR(255) NOT NULL,
  mime_type VARCHAR(100),
  size_bytes BIGINT NOT NULL,

  minio_bucket VARCHAR(255) NOT NULL,
  minio_key VARCHAR(1024) NOT NULL,

  status VARCHAR(20) NOT NULL DEFAULT 'UPLOADED',
  classification VARCHAR(20) NOT NULL DEFAULT 'PUBLIC',
  retention_days INT,

  current_version INT NOT NULL DEFAULT 1,
  total_versions INT NOT NULL DEFAULT 1,

  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- File versions table
CREATE TABLE document.file_versions (
  version_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  file_id UUID NOT NULL REFERENCES document.files(file_id) ON DELETE CASCADE,
  version_number INT NOT NULL,
  minio_key VARCHAR(1024) NOT NULL,
  size_bytes BIGINT NOT NULL,
  created_by UUID NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Idempotency table for NATS consumer
CREATE TABLE document.processed_events (
  event_id UUID PRIMARY KEY,
  subject VARCHAR(255) NOT NULL,
  processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_files_tenant_created ON document.files (tenant_id, created_at DESC);
CREATE INDEX idx_files_entity ON document.files (entity_type, entity_id);
CREATE INDEX idx_files_status ON document.files (status);
CREATE INDEX idx_files_classification ON document.files (classification);
CREATE INDEX idx_versions_file_id ON document.file_versions (file_id, version_number DESC);
CREATE INDEX idx_processed_events_subject ON document.processed_events (subject);

-- Enable RLS on files table
ALTER TABLE document.files ENABLE ROW LEVEL SECURITY;
ALTER TABLE document.files FORCE ROW LEVEL SECURITY;

-- RLS policy for tenant isolation
CREATE POLICY tenant_isolation ON document.files
  USING (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

-- Grant permissions to application role
GRANT SELECT, INSERT, UPDATE ON document.files TO hris_app;
GRANT SELECT, INSERT ON document.file_versions TO hris_app;
GRANT SELECT, INSERT ON document.processed_events TO hris_app;
