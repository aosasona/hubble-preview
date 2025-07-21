CREATE TYPE entry_status AS ENUM (
	'queued',
	'processing',
	'completed',
	'failed',
	'canceled',
	'paused'
);

CREATE TABLE IF NOT EXISTS entries_queue (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

	entry_id INT NOT NULL REFERENCES entries(id),
	payload JSONB NOT NULL, -- additional data for the worker

	status entry_status NOT NULL DEFAULT 'queued',
	attempts INT NOT NULL DEFAULT 0,
	max_attempts INT NOT NULL DEFAULT 3,

	available_at TIMESTAMP NOT NULL DEFAULT now(), -- for exponential backoff retries
	created_at TIMESTAMP NOT NULL DEFAULT now(),
	updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_entries_queue_entry_id ON entries_queue(entry_id);
CREATE INDEX IF NOT EXISTS idx_entries_queue_status_available_at ON entries_queue(status, available_at);

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON entries_queue
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

