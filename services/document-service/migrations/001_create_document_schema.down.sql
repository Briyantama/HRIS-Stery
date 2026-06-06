-- Drop tables (order matters due to foreign key)
DROP TABLE IF EXISTS document.file_versions;
DROP TABLE IF EXISTS document.files;
DROP TABLE IF EXISTS document.processed_events;

-- Drop schema
DROP SCHEMA IF EXISTS document;
