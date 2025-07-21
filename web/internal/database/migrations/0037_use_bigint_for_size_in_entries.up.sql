ALTER TABLE entries DROP COLUMN size;

ALTER TABLE entries ADD COLUMN filesize_bytes BIGINT NOT NULL DEFAULT 0;

COMMENT ON COLUMN entries.filesize_bytes IS 'Size of the file in bytes';

