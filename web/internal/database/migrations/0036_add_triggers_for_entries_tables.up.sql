CREATE TRIGGER set_updated_at
BEFORE UPDATE ON entries
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON entry_chunks
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

