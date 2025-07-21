CREATE TYPE entry_chunk_embedding_status AS ENUM ('pending', 'done', 'failed');

ALTER TABLE entry_chunks
ADD COLUMN IF NOT EXISTS embedding_status entry_chunk_embedding_status NOT NULL DEFAULT 'pending',
ADD COLUMN IF NOT EXISTS embedding_status_updated_at TIMESTAMPTZ DEFAULT NULL,
ADD COLUMN IF NOT EXISTS last_embedding_error TEXT,
ADD COLUMN IF NOT EXISTS last_embedding_error_at TIMESTAMPTZ DEFAULT NULL,
ADD COLUMN IF NOT EXISTS embedding_error_count INT DEFAULT 0;

-- Add a trigger to update the embedding status and error count
create or replace function update_embedding_status()
returns trigger
as $$
BEGIN
	IF NEW.embedding_status = 'done' THEN
		NEW.embedding_status_updated_at = NOW();
		NEW.embedding_error_count = 0;
	ELSIF NEW.embedding_status = 'failed' THEN
		NEW.last_embedding_error = NEW.content; -- Assuming content contains the error message
		NEW.last_embedding_error_at = NOW();
		NEW.embedding_error_count = COALESCE(NEW.embedding_error_count, 0) + 1;
	END IF;
	RETURN NEW;
END;
$$
language plpgsql
;

CREATE TRIGGER trg_update_embedding_status
BEFORE INSERT OR UPDATE OF embedding_status ON entry_chunks
FOR EACH ROW
EXECUTE FUNCTION update_embedding_status();

